package main

import (
	"github.com/Jailior/open-search/backend/internal/crawler"
	"github.com/Jailior/open-search/backend/internal/models"
	"github.com/Jailior/open-search/backend/internal/storage"
)

const DB_NAME = "opensearch"
const COLL_NAME = "pages"

/*
Initializes a URL queue for BFS crawling and a visited set
to avoid repeating web pages.
Seeds the crawler with a single initial seed.
*/
func main() {

	db := storage.MakeDB()
	db.Connect()
	db.AddCollection(DB_NAME, COLL_NAME, crawler.PAGE_INSERT_COLLECTION)
	defer db.Disconnect()

	queue := models.MakeURLQueue()
	visited := models.MakeSet()

	seed := "https://ubc.ca"
	// seed := "https://www.magazine.alumni.ubc.ca/2025/campus-community/meet-3-ubcs-oldest-surviving-clubs"
	crawler.StartCrawler(seed, queue, visited, db)
}
