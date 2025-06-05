package stats

import (
	"context"
	"sync"
	"time"

	"github.com/Jailior/open-search/backend/internal/storage"
	"github.com/Jailior/open-search/backend/internal/utils"
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

				_ = utils.RetryWithBackoff(func() error {
					_, err := db.GetCollection("pages").UpdateOne(context.Background(), filter, update, opts)
					return err
				}, 3, "UpdateCrawlerStats")

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

func (stats *CrawlerStats) TrackQueueSize(rdb *storage.RedisClient) {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				qLength, _ := rdb.Client.LLen(rdb.Ctx, "url_queue").Result()
				stats.mu.Lock()
				stats.currentQLength = int(qLength)
				stats.mu.Unlock()
			case <-stats.stopChan:
				return
			}
		}
	}()
}

// Updates cumulative number of pages visited and queue size
func (stats *CrawlerStats) PageVisit() {
	stats.mu.Lock()
	defer stats.mu.Unlock()
	stats.currentPagesVisited++
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
