package crawler

import (
	"fmt"
	"log"
	"time"

	"github.com/Jailior/open-search/backend/internal/models"
	"github.com/Jailior/open-search/backend/internal/parsing"
	"github.com/Jailior/open-search/backend/internal/storage"
	"github.com/PuerkitoBio/goquery"
	"github.com/abadojack/whatlanggo"
	"github.com/gocolly/colly/v2"
)

const maxChars = 100_000
const lang_sample_size = 100

// Crawler context passed to HTML handler and others
type CrawlContext struct {
	Queue      *models.URLQueue
	VisitedSet *models.Set
	Database   *storage.Database
	Stats      *models.CrawlerStats
}

func StartCrawler(seedURL string, q *models.URLQueue, visited *models.Set, db *storage.Database) {
	var err error
	pagesVisited := 0

	// intial colly scraper
	c := colly.NewCollector()
	err = c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 2,
		Delay:       2 * time.Second,
		RandomDelay: 1 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}

	// initialize statistics struct
	stats := models.MakeCrawlerStats()

	// initialize html handler for processing page contents
	handler := MakeHTMLHandler(&CrawlContext{
		Queue:      q,
		VisitedSet: visited,
		Database:   db,
		Stats:      stats,
	})

	// Processes page contents
	c.OnHTML("html", handler)

	// enqueue seedURL
	q.Enqueue(seedURL)

	// bfs crawling
	for !q.IsEmpty() {
		url, ok := q.Dequeue()
		if !ok {
			break
		}
		if visited.Has(url) {
			stats.DuplicatesAvoided++
			continue
		}
		c.Visit(url)
		visited.Add(url)
		pagesVisited++
		stats.Update(pagesVisited, q.Length())
	}
}

func MakeHTMLHandler(ctx *CrawlContext) func(e *colly.HTMLElement) {

	return func(e *colly.HTMLElement) {
		// reassign context variables
		q := ctx.Queue
		db := ctx.Database
		stats := ctx.Stats

		// clean url
		var err error
		url := e.Request.AbsoluteURL(e.Request.URL.String())
		url, err = parsing.NormalizeAndStripURL(url)
		if err != nil {
			fmt.Println("Failed to parse URL:", err)
			stats.PagesSkippedErr++
			return
		}

		title := e.DOM.Find("title").Text()

		// extract and clean
		doc := e.DOM
		doc.Find("script, style, noscript, iframe, nav, footer, header, form, link").Each(func(i int, s *goquery.Selection) {
			s.Remove()
		})
		content := parsing.CleanText(doc)

		// detect language
		if len(content) < lang_sample_size {
			stats.PagesSkippedErr++
			return // page too short
		}
		sample := content[:lang_sample_size]
		info := whatlanggo.Detect(sample)
		if info.Lang != whatlanggo.Eng {
			stats.PagesSkippedLang++
			return // don't process non-English pages
		}

		// limit page size
		if len(content) > maxChars {
			content = content[:maxChars]
		}

		// collect outlinks
		var outlinks []string
		e.DOM.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
			href, ok := s.Attr("href")
			if !ok {
				return
			}
			abs_href := e.Request.AbsoluteURL(href)
			abs_href, err = parsing.NormalizeAndStripURL(abs_href)
			if err != nil {
				fmt.Println("Failed to parse URL:", err)
				stats.PagesSkippedErr++
				return
			}
			if abs_href != "" {
				outlinks = append(outlinks, abs_href)
				q.Enqueue(abs_href)
			}
		})

		// store page
		page := models.PageData{
			Title:       title,
			URL:         url,
			Content:     content,
			Outlinks:    outlinks,
			TimeCrawled: time.Now(),
		}

		err = db.InsertPage(&page)
		if err != nil {
			panic(err)
		}

		models.PrintPage(page)
	}
}
