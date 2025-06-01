package stats

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/Jailior/open-search/backend/internal/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Crawler statistics, thread-safe
type CrawlerStats struct {
	PagesCrawled      []uint32      `bson:"pages_crawled"`
	QueueSize         []int         `bson:"queue_size"`
	PagesSkippedErr   int           `bson:"page_errs"`
	PagesSkippedLang  int           `bson:"pages_skipped_lang"`
	DuplicatesAvoided int           `bson:"duplicates_avoided"`
	LastUpdated       time.Time     `bson:"-"`
	mu                sync.Mutex    `bson:"-"`
	stopChan          chan struct{} `bson:"-"`

	currentQLength      int
	currentPagesVisited int
}

// Returns default instance of CrawlerStats
func MakeCrawlerStats() *CrawlerStats {
	return &CrawlerStats{
		PagesCrawled:        make([]uint32, 0),
		QueueSize:           make([]int, 0),
		PagesSkippedErr:     0,
		PagesSkippedLang:    0,
		DuplicatesAvoided:   0,
		LastUpdated:         time.Now(),
		stopChan:            make(chan struct{}),
		currentQLength:      0,
		currentPagesVisited: 0,
	}
}

// Background writer, writes to filename every interval with updated JSON object for crawler stats
func (stats *CrawlerStats) StartWriter(interval time.Duration, db *storage.Database) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				stats.mu.Lock()

				stats.PagesCrawled = append(stats.PagesCrawled, uint32(stats.currentPagesVisited))
				stats.QueueSize = append(stats.QueueSize, stats.currentQLength)
				stats.LastUpdated = time.Now()

				saved_stats := *stats
				stats.mu.Unlock()

				filter := bson.M{"_id": "crawler_stats"}
				update := bson.M{"$set": saved_stats}

				opts := options.Update().SetUpsert(true)

				_, err := db.GetCollection("pages").UpdateOne(context.Background(), filter, update, opts)

				if err != nil {
					log.Println("Failed to write CrawlerStats to MongoDB:", err)
					continue
				}

			case <-stats.stopChan:
				return
			}

		}
	}()
}

// Stops writer
func (stats *CrawlerStats) StopWriter() {
	close(stats.stopChan)
}

// Updates cumulative number of pages visited and queue size
func (stats *CrawlerStats) PageVisit(queueLength int) {
	stats.mu.Lock()
	defer stats.mu.Unlock()
	stats.currentPagesVisited++
	stats.currentQLength = queueLength
}

// Thread-safe incrementers

func (stats *CrawlerStats) IncrementSkippedErr() {
	stats.mu.Lock()
	defer stats.mu.Unlock()
	stats.PagesSkippedErr++
}

func (stats *CrawlerStats) IncrementSkippedLang() {
	stats.mu.Lock()
	defer stats.mu.Unlock()
	stats.PagesSkippedLang++
}

func (stats *CrawlerStats) IncrementSkippedDupe() {
	stats.mu.Lock()
	defer stats.mu.Unlock()
	stats.DuplicatesAvoided++
}
