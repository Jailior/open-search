package main

import (
	"github.com/Jailior/ubc-search/backend/internal/crawler"
	"github.com/Jailior/ubc-search/backend/internal/models"
)

/*
Initializes a URL queue for BFS crawling and a visited set
to avoid repeating web pages.
Seeds the crawler with a single initial seed.
*/
func main() {
	queue := models.MakeURLQueue()
	visited := models.MakeSet()
	seed := "https://en.wikipedia.org/wiki/Sobel_operator"
	crawler.StartCrawler(seed, queue, visited)
}
