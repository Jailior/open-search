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
	"github.com/PuerkitoBio/goquery"
	"github.com/abadojack/whatlanggo"
	"github.com/gocolly/colly/v2"
)

const maxChars = 100_000
const lang_sample_size = 100

const DB_NAME = "opensearch"
const PAGE_INSERT_COLLECTION = "pages"

const REDIS_INDEX_QUEUE = "pages_to_index"

const REDIS_URL_QUEUE = "url_queue"
const REDIS_VISITED_SET = "visited_set"

// Crawler context passed to HTML handler and others
type CrawlContext struct {
	Database *storage.Database
	Stats    *stats.CrawlerStats
	Redis    *storage.RedisClient
	Err      error
}

func StartCrawler(seeds []string, ctx *CrawlContext, workerCount int, cancelContext context.Context) {
	// enqueue seed URLs
	for _, url := range seeds {
		ctx.Redis.EnqueueList(url, REDIS_URL_QUEUE, REDIS_VISITED_SET)
	}

	length, _ := ctx.Redis.Client.LLen(ctx.Redis.Ctx, REDIS_URL_QUEUE).Result()
	fmt.Println("URL queue length:", length)

	time.Sleep(100 * time.Millisecond)

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

func runCrawler(ctx *CrawlContext, workerID int, cancelCtx context.Context) {
	var err error

	stats := ctx.Stats
	rdb := ctx.Redis

	// intial colly scraper
	c := colly.NewCollector()
	err = c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 1,
		Delay:       0 * time.Second,
		RandomDelay: 0 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}

	// initialize html handler, called on html visit
	handler := MakeHTMLHandler(ctx)

	// Processes page contents
	c.OnHTML("html", handler)

	log.Printf("[Worker %d] started\n", workerID)

	// bfs crawling
	for {
		select {
		// shutdown
		case <-cancelCtx.Done():
			log.Printf("Worker [%d] received shutdown signal\n", workerID)
			return
		default:
			// dequeue
			url, err := rdb.DequeueList(REDIS_URL_QUEUE)
			if err != nil {
				log.Printf("[Worker %d] Queue pop error: %v\n", workerID, err)
				continue
			}

			// check for duplicates
			if rdb.SetHas(url, REDIS_VISITED_SET) {
				stats.IncrementSkippedDupe()
				continue
			}

			// visit page
			err = c.Visit(url)
			if err != nil {
				log.Printf("[Worker %d] Failed to visit page.\n", workerID)
				continue
			}
			if ctx.Err == nil {
				stats.PageVisit()
			}
			ctx.Err = nil

			// log stats
			rdb.SetAdd(url, REDIS_VISITED_SET)
		}
	}
}

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
				log.Println("Failed to parse URL:", err)
				stats.IncrementSkippedErr()
				return
			}
			if abs_href != "" {
				outlinks = append(outlinks, abs_href)
				if !rdc.SetHas(abs_href, REDIS_VISITED_SET) {
					rdc.EnqueueList(abs_href, REDIS_URL_QUEUE, REDIS_VISITED_SET)
				} else {
					stats.IncrementSkippedDupe()
				}
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
			if strings.Contains(err.Error(), "DUPE") {
				stats.IncrementSkippedDupe()
				log.Printf("Database Deduplication raised and handled: %v", err)
			} else {
				ctx.Err = err
			}
			return
		}

		// push page id to redis stream
		err = rdc.PushToStream(REDIS_INDEX_QUEUE, "id", id)
		if err != nil {
			log.Println("Failed to push to Redis stream: ", err)
			ctx.Err = err
			return
		}

		models.PrintPage(page)
	}
}
