package models

import (
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

/* Crawler Models*/

// Key data points stored per page
type PageData struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	URL         string             `bson:"url"`
	Title       string             `bson:"title"`
	Content     string             `bson:"content"`
	Outlinks    []string           `bson:"outlinks"`
	TimeCrawled time.Time          `bson:"timecrawled"`
}

// Prints key elements of a page
func PrintPage(pd PageData) {
	log.Println("Title: ", pd.Title)
	log.Println("URL: ", pd.URL)
	// log.Println("# of Outlinks: ", len(pd.Outlinks))
}

/* Indexer Models */

// Redis Indexer Queue Message type
type IndexerQueueMessage struct {
	ID string `json:"id"` // mongodb hex _id string
}

// Information stored for a posting in a term document
type IndexerPosting struct {
	DocID     string  `bson:"doc_id"`
	Title     string  `bson:"title"`
	URL       string  `bson:"url"`
	TF        float64 `bson:"TF"`
	Positions []int   `bson:"positions"`
}

// Information representing a term document in database
type TermEntry struct {
	DF       int              `bson:"df"`
	Postings []IndexerPosting `bson:"postings"`
}

/* PageRank models */

// PageRank document stored in collection
type PageRankScore struct {
	URL   string  `bson:"url"`
	Score float64 `bson:"score"`
}

/* Deprecated Models */

// // Thread-safe FIFO Queue
// type URLQueue struct {
// 	elems []string
// 	mux   sync.Mutex
// }

// // Returns an empty URLQueue
// func MakeURLQueue() *URLQueue {
// 	return &URLQueue{
// 		elems: make([]string, 0),
// 	}
// }

// // Adds url to end of queue
// func (q *URLQueue) Enqueue(url string) {
// 	q.mux.Lock()
// 	defer q.mux.Unlock()

// 	q.elems = append(q.elems, url)
// }

// // Pops off and returns first item in queue
// // Returns "" and false if not elements left in queue
// func (q *URLQueue) Dequeue() (string, bool) {
// 	q.mux.Lock()
// 	defer q.mux.Unlock()

// 	if len(q.elems) == 0 {
// 		return "", false
// 	}

// 	elem := q.elems[0]
// 	q.elems = q.elems[1:]
// 	return elem, true
// }

// // Returns true if queue is empty
// func (q *URLQueue) IsEmpty() bool {
// 	q.mux.Lock()
// 	defer q.mux.Unlock()
// 	return len(q.elems) == 0
// }

// // Returns the length of the queue
// func (q *URLQueue) Length() int {
// 	q.mux.Lock()
// 	defer q.mux.Unlock()
// 	return len(q.elems)
// }

// // Visited Set, avoiding repeat visits, thread-safe
// type Set struct {
// 	set map[string]struct{}
// 	mux sync.Mutex
// }

// // Returns a new empty Set
// func MakeSet() *Set {
// 	return &Set{
// 		set: make(map[string]struct{}),
// 	}
// }

// // Adds url to set
// func (s *Set) Add(url string) {
// 	s.mux.Lock()
// 	defer s.mux.Unlock()
// 	s.set[url] = struct{}{}
// }

// // Returns true if url is in the set
// func (s *Set) Has(url string) bool {
// 	s.mux.Lock()
// 	defer s.mux.Unlock()
// 	_, ok := s.set[url]
// 	return ok
// }

// // No need for Remove function
