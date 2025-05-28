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
const PAGE_INSERT_COLLECTION = "raw_pages"

// Crawler context passed to HTML handler and others
type CrawlContext struct {
	Queue      *models.URLQueue
	VisitedSet *models.Set
	Database   *storage.Database
	Stats      *models.CrawlerStats
	Redis      *storage.RedisClient
	Err        error
}

func StartCrawler(seedURL string, q *models.URLQueue, visited *models.Set, db *storage.Database) {
	var err error

	// intial colly scraper
	c := colly.NewCollector()
	err = c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 2,
		Delay:       0 * time.Second,
		RandomDelay: 0 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}

	// initialize statistics struct
	stats := models.MakeCrawlerStats()
	stats.StartWriter(1*time.Minute, "stats.json")
	defer stats.StopWriter()

	// initialize redis client
	rdc := storage.MakeRedisClient()

	// initialize html handler and crawl context for processing page contents
	ctx := &CrawlContext{
		Queue:      q,
		VisitedSet: visited,
		Database:   db,
		Stats:      stats,
		Redis:      rdc,
	}
	handler := MakeHTMLHandler(ctx)

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
			stats.IncrementSkippedDupe()
			continue
		}
		c.Visit(url)
		visited.Add(url)
		if ctx.Err == nil {
			stats.PageVisit(q.Length())
		}
	}
}

func MakeHTMLHandler(ctx *CrawlContext) func(e *colly.HTMLElement) {

	return func(e *colly.HTMLElement) {
		// reassign context variables
		q := ctx.Queue
		db := ctx.Database
		stats := ctx.Stats
		rdc := ctx.Redis

		// erase previous errors
		ctx.Err = nil

		// clean url
		var err error
		url := e.Request.AbsoluteURL(e.Request.URL.String())
		url, err = parsing.NormalizeAndStripURL(url)
		if err != nil {
			fmt.Println("Failed to parse URL:", err)
			stats.IncrementSkippedErr()
			ctx.Err = err
			return
		}

		title := e.DOM.Find("title").Text()

		// extract and clean
		doc := e.DOM
		content := parsing.CleanText(doc)

		// detect language
		if len(content) < lang_sample_size {
			stats.IncrementSkippedErr()
			ctx.Err = fmt.Errorf("Page too short error.")
			return // page too short
		}
		sample := content[:lang_sample_size]
		info := whatlanggo.Detect(sample)
		if info.Lang != whatlanggo.Eng {
			stats.IncrementSkippedLang()
			ctx.Err = fmt.Errorf("Non-English page skipped.")
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
				stats.IncrementSkippedErr()
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

		id, err := db.InsertRawPage(&page, PAGE_INSERT_COLLECTION)
		if err != nil {
			ctx.Err = err
			panic(err)
		}

		err = rdc.PushToStream("pages_to_index", "id", id)
		if err != nil {
			fmt.Println("Failed to push to Redis stream: ", err)
			ctx.Err = err
			return
		}

		models.PrintPage(page)
	}
}
