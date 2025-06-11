package crawler

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/Jailior/open-search/backend/internal/models"
	"github.com/Jailior/open-search/backend/internal/parsing"
	"github.com/Jailior/open-search/backend/internal/stats"
	"github.com/Jailior/open-search/backend/internal/storage"
	"github.com/Jailior/open-search/backend/internal/utils"
	"github.com/PuerkitoBio/goquery"
	"github.com/abadojack/whatlanggo"
	"github.com/gocolly/colly/v2"
)

// Maximum number of characters in a page
const maxChars = 100_000

// Number of characters used for detecting language of page
const lang_sample_size = 100

// MongoDB database name
const DB_NAME = "opensearch"

// Collection name raw pages are inserted into
const PAGE_INSERT_COLLECTION = "pages"

// Redis stream name to send DocIDs of pages to be indexed
const REDIS_INDEX_QUEUE = "pages_to_index"

// Redis list name of url queue
const REDIS_URL_QUEUE = "url_queue"

// Redis set name of visited set
const REDIS_VISITED_SET = "visited_set"

// Crawler context passed to HTML handler and others
type CrawlContext struct {
	Database *storage.Database
	Stats    *stats.CrawlerStats
	Redis    *storage.RedisClient
	Err      error
}

// Starts N crawlers with a crawl context and a background context
func StartCrawler(ctx *CrawlContext, workerCount int, cancelContext context.Context) {

	// Get length of URL queue
	length, _ := ctx.Redis.Client.LLen(ctx.Redis.Ctx, REDIS_URL_QUEUE).Result()
	fmt.Println("URL queue length:", length)

	time.Sleep(100 * time.Millisecond)

	// Initialize workers
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			runCrawler(ctx, workerID, cancelContext)
		}(i)
	}

	wg.Wait()
	log.Println("All workers shut down.")
}

// Crawls pages in the URL queue until background context sends a shutdown signal
func runCrawler(ctx *CrawlContext, workerID int, cancelCtx context.Context) {
	var err error
	stats := ctx.Stats
	rdb := ctx.Redis

	// intial colly scraper
	c := colly.NewCollector(
		colly.DisallowedDomains(
			"https://redditinc.com/", // causes slow parsing
		),
	)
	// Crawl limiters
	err = c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 1,
		Delay:       0 * time.Second,
		RandomDelay: 0 * time.Second,
	})
	// Failed to initialize limiters
	if err != nil {
		log.Fatal(err)
	}

	// initialize html handler, called on html visit
	handler := MakeHTMLHandler(ctx)

	// Processes page contents upon html visit
	c.OnHTML("html", handler)

	log.Printf("[Worker %d] started\n", workerID)

	// bfs crawling
	for {
		// Selects between crawling next page, default, and shutting down if signal is sent
		select {
		// shutdown
		case <-cancelCtx.Done():
			log.Printf("Worker [%d] received shutdown signal\n", workerID)
			return
		default:
			// default behaviour

			// dequeue urql from Redis queue
			url, err := rdb.DequeueList(REDIS_URL_QUEUE)
			if err != nil {
				log.Printf("[Worker %d] Queue pop error: %v\n", workerID, err)
				continue
			}

			// check if page is already in visited, if yes skip it
			if rdb.SetHas(url, REDIS_VISITED_SET) {
				stats.IncrementSkippedDupe()
				continue
			}

			// "visit" page, initializes html handler on page
			err = c.Visit(url)

			// if an error occured in the html handler
			if err != nil {
				log.Printf("[Worker %d] Failed to visit page.\n", workerID)
				continue
			}
			// if no errors were logged, increment stats metrics
			if ctx.Err == nil {
				stats.PageVisit()
			}
			ctx.Err = nil

			// add URL to Redis visited set
			rdb.SetAdd(url, REDIS_VISITED_SET)
		}
	}
}

// HTML handler maker, used to make a function signature compliant handler
// while providing the crawl context
func MakeHTMLHandler(ctx *CrawlContext) func(e *colly.HTMLElement) {

	return func(e *colly.HTMLElement) {
		// reassign context variables
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
			log.Println("Failed to parse URL:", err)
			stats.IncrementSkippedErr()
			ctx.Err = err
			return
		}

		// find the page title
		title := e.DOM.Find("title").Text()

		// extract and clean
		doc := e.DOM
		content := parsing.CleanText(doc)

		// detect language
		// if page is too short for a language detection, skip it
		if len(content) < lang_sample_size {
			stats.IncrementSkippedErr()
			ctx.Err = fmt.Errorf("Page too short error.")
			return // page too short
		}
		// get language sample
		sample := content[:lang_sample_size]

		// detect language via whatlanggo
		info := whatlanggo.Detect(sample)
		// if page is not in English skip it
		if info.Lang != whatlanggo.Eng {
			stats.IncrementSkippedLang()
			ctx.Err = fmt.Errorf("Non-English page skipped.")
			return // don't process non-English pages
		}

		// limit page size
		if len(content) > maxChars {
			content = content[:maxChars]
		}

		// collect outlinks, and add them to Redis queue
		var outlinks []string

		// For Each idiom with function called on every a[href] in page
		e.DOM.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
			// get link
			href, ok := s.Attr("href")
			if !ok {
				return
			}

			// get absolute URL
			abs_href := e.Request.AbsoluteURL(href)
			// normalize and strip URL
			abs_href, err = parsing.NormalizeAndStripURL(abs_href)
			// if parsing URL caused an error, skip it
			if err != nil {
				log.Println("Failed to parse URL:", err)
				stats.IncrementSkippedErr()
				return
			}
			// if the link is valid then add it
			if abs_href != "" {
				// append to outlinks to be added to Mongo document
				outlinks = append(outlinks, abs_href)

				// if the visited includes the page referenced by URL, then don't add to queue
				if !rdc.SetHas(abs_href, REDIS_VISITED_SET) {

					// add to the queue with 3 retries on error
					err = utils.RetryWithBackoff(func() error {
						return rdc.EnqueueList(abs_href, REDIS_URL_QUEUE, REDIS_VISITED_SET)
					}, 3, "Redis-EnqueueList")

				} else {
					// increment stats skipped dupe
					stats.IncrementSkippedDupe()
				}
			}
		})

		// make page instance
		page := models.PageData{
			Title:       title,
			URL:         url,
			Content:     content,
			Outlinks:    outlinks,
			TimeCrawled: time.Now(),
		}

		// insert raw page into database
		id, err := db.InsertRawPage(&page, PAGE_INSERT_COLLECTION)
		if err != nil {
			// Detect if duplicate page is being added
			if strings.Contains(err.Error(), "DUPE") {
				stats.IncrementSkippedDupe()
				log.Printf("Database Deduplication raised and handled: %v", err)
			} else {
				// any other error
				ctx.Err = err
				stats.IncrementSkippedErr()
			}
			// return on error, i.e. don't add to redis queue
			return
		}

		// push page id to redis stream, with 3 retries
		err = utils.RetryWithBackoff(func() error {
			e := rdc.PushToStream(REDIS_INDEX_QUEUE, "id", id)
			return e
		}, 3, "Redis-StreamPush")

		// check for Redis stream error
		if err != nil {
			log.Println("Failed to push to Redis stream: ", err)
			ctx.Err = err
			return
		}

		// Print page for logging
		models.PrintPage(page)
	}
}
