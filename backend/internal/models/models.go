package models

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Key data points stored per page
type PageData struct {
	URL         string
	Title       string
	Content     string
	Outlinks    []string
	TimeCrawled time.Time
}

// Prints key elements of a page
func PrintPage(pd PageData) {
	fmt.Println("Title: ", pd.Title)
	fmt.Println("URL: ", pd.URL)
	fmt.Println("# of Outlinks: ", len(pd.Outlinks))
}

// Crawler statistics
type CrawlerStats struct {
	PagesCrawled      []uint32
	QueueSize         []int
	PagesSkippedErr   int
	PagesSkippedLang  int
	DuplicatesAvoided int
	LastUpdated       time.Time
}

func MakeCrawlerStats() *CrawlerStats {
	return &CrawlerStats{
		PagesCrawled:      make([]uint32, 0),
		QueueSize:         make([]int, 0),
		PagesSkippedErr:   0,
		PagesSkippedLang:  0,
		DuplicatesAvoided: 0,
		LastUpdated:       time.Now(),
	}
}

func (stats *CrawlerStats) Update(pagesVisited int, qLength int) {
	if time.Now().Sub(stats.LastUpdated) >= time.Minute {
		stats.LastUpdated = time.Now()
		stats.PagesCrawled = append(stats.PagesCrawled, uint32(pagesVisited))
		stats.QueueSize = append(stats.QueueSize, qLength)

		// save as json
		jsonData, jsonErr := json.Marshal(stats)
		if jsonErr != nil {
			fmt.Println("Error creating json object", jsonErr)
		}
		file, fileError := os.Create("stats.json")
		if fileError != nil {
			fmt.Println("Error creating file", fileError)
		}
		defer file.Close()
		_, writeError := file.Write(jsonData)
		if writeError != nil {
			fmt.Println("Error writing to file", writeError)
		}
	}
}

// Thread-safe FIFO Queue
type URLQueue struct {
	elems []string
	mux   sync.Mutex
}

// Returns an empty URLQueue
func MakeURLQueue() *URLQueue {
	return &URLQueue{
		elems: make([]string, 0),
	}
}

// Adds url to end of queue
func (q *URLQueue) Enqueue(url string) {
	q.mux.Lock()
	defer q.mux.Unlock()

	q.elems = append(q.elems, url)
}

// Pops off and returns first item in queue
// Returns "" and false if not elements left in queue
func (q *URLQueue) Dequeue() (string, bool) {
	q.mux.Lock()
	defer q.mux.Unlock()

	if len(q.elems) == 0 {
		return "", false
	}

	elem := q.elems[0]
	q.elems = q.elems[1:]
	return elem, true
}

// Returns true if queue is empty
func (q *URLQueue) IsEmpty() bool {
	q.mux.Lock()
	defer q.mux.Unlock()
	return len(q.elems) == 0
}

// Returns the length of the queue
func (q *URLQueue) Length() int {
	q.mux.Lock()
	defer q.mux.Unlock()
	return len(q.elems)
}

// Visited Set, avoiding repeat visits, thread-safe
type Set struct {
	set map[string]struct{}
	mux sync.Mutex
}

// Returns a new empty Set
func MakeSet() *Set {
	return &Set{
		set: make(map[string]struct{}),
	}
}

// Adds url to set
func (s *Set) Add(url string) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.set[url] = struct{}{}
}

// Returns true if url is in the set
func (s *Set) Has(url string) bool {
	s.mux.Lock()
	defer s.mux.Unlock()
	_, ok := s.set[url]
	return ok
}

// No need for Remove function
