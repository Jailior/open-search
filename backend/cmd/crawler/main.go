package main

import (
	"github.com/Jailior/open-search/backend/internal/crawler"
	"github.com/Jailior/open-search/backend/internal/models"
	"github.com/Jailior/open-search/backend/internal/storage"
)

/*
Initializes a URL queue for BFS crawling and a visited set
to avoid repeating web pages.
Seeds the crawler with a single initial seed.
*/
func main() {

	db := storage.MakeDB()
	db.Connect()
	defer db.Disconnect()

	queue := models.MakeURLQueue()
	visited := models.MakeSet()

	seed := "https://en.wikipedia.org/wiki/Sobel_operator"
	crawler.StartCrawler(seed, queue, visited, db)
}
