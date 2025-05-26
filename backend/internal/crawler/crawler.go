package crawler

import (
	"strings"

	"github.com/Jailior/ubc-search/internal/models"
	"github.com/PuerkitoBio/goquery"
	"github.com/abadojack/whatlanggo"
	"github.com/gocolly/colly/v2"
)

const maxChars = 100_000
const lang_sample_size = 100

func StartCrawler(seedURL string, q *models.URLQueue, visited *models.Set) {
	c := colly.NewCollector()

	// Add more advanced for checking where this link points to
	// Processes page contents
	c.OnHTML("html", func(e *colly.HTMLElement) {
		url := e.Request.AbsoluteURL(e.Request.URL.String())
		title := e.DOM.Find("title").Text()

		// extract and clean
		doc := e.DOM
		doc.Find("script, style, noscript, iframe, nav, footer, header, form, link").Each(func(i int, s *goquery.Selection) {
			s.Remove()
		})
		text := doc.Find("body").Text()
		content := strings.TrimSpace(text)
		content = strings.Join(strings.Fields(content), " ")

		// detect language
		if len(content) < lang_sample_size {
			return // page too short
		}
		sample := content[:lang_sample_size]
		info := whatlanggo.Detect(sample)
		if info.Lang != whatlanggo.Eng {
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
			if abs_href != "" {
				outlinks = append(outlinks, abs_href)
				q.Enqueue(abs_href)
			}
		})

		// store page
		page := models.PageData{
			Title:    title,
			URL:      url,
			Content:  content,
			Outlinks: outlinks,
		}

		models.PrintPage(page)
	})

	// q.Enqueue(seedURL)

	c.Visit(seedURL)

	// for !q.IsEmpty() {
	// 	url, ok := q.Dequeue()
	// 	if !ok {
	// 		break
	// 	}
	// 	if visited.Has(url) {
	// 		continue
	// 	}
	// 	c.Visit(url)
	// 	visited.Add(url)
	// }
}
