package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Jailior/open-search/backend/internal/crawler"
	"github.com/Jailior/open-search/backend/internal/stats"
	"github.com/Jailior/open-search/backend/internal/storage"
)

const NUM_WORKERS = 4

/*
Initializes a URL queue for BFS crawling and a visited set
to avoid repeating web pages.
Seeds the crawler with a single initial seed.
*/
func main() {

	// shutdown context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// capture interrupt or term signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		log.Println("Shutdown signal received...")
		cancel()
	}()

	// initialize database connection
	db := storage.MakeDB()
	db.Connect()
	db.AddCollection(crawler.DB_NAME, crawler.PAGE_INSERT_COLLECTION)
	defer db.Disconnect()

	// initialize redis client
	rdc := storage.MakeRedisClient()
	rdc.ResetQueueAndSet(crawler.REDIS_URL_QUEUE, crawler.REDIS_VISITED_SET)

	// initialize statistics struct
	stats := stats.MakeCrawlerStats()
	stats.StartWriter(1*time.Minute, db)
	stats.TrackQueueSize(rdc)
	defer stats.StopWriter()

	seeds := make([]string, 0)

	seeds = append(seeds, "https://www.google.com/")
	seeds = append(seeds, "https://www.yahoo.com/")
	seeds = append(seeds, "https://www.reddit.com/")

	// initialize crawl context
	crawlCtx := &crawler.CrawlContext{
		Database: db,
		Stats:    stats,
		Redis:    rdc,
	}

	crawler.StartCrawler(seeds, crawlCtx, NUM_WORKERS, ctx)
}
