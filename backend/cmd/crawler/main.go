package main

import (
	"time"

	"github.com/Jailior/open-search/backend/internal/crawler"
	"github.com/Jailior/open-search/backend/internal/models"
	"github.com/Jailior/open-search/backend/internal/stats"
	"github.com/Jailior/open-search/backend/internal/storage"
)

/*
Initializes a URL queue for BFS crawling and a visited set
to avoid repeating web pages.
Seeds the crawler with a single initial seed.
*/
func main() {

	// initialize database connection
	db := storage.MakeDB()
	db.Connect()
	db.AddCollection(crawler.DB_NAME, crawler.PAGE_INSERT_COLLECTION)
	defer db.Disconnect()

	// initialize statistics struct
	stats := stats.MakeCrawlerStats()
	stats.StartWriter(1*time.Minute, db)
	defer stats.StopWriter()

	// initialize queue and set
	queue := models.MakeURLQueue()
	visited := models.MakeSet()

	seed := "https://ubc.ca"
	// seed := "https://www.magazine.alumni.ubc.ca/2025/campus-community/meet-3-ubcs-oldest-surviving-clubs"

	// initialize redis client
	rdc := storage.MakeRedisClient()

	// initialize crawl context
	ctx := &crawler.CrawlContext{
		Queue:      queue,
		VisitedSet: visited,
		Database:   db,
		Stats:      stats,
		Redis:      rdc,
	}

	crawler.StartCrawler(seed, ctx)
}
