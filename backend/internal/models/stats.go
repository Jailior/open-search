package models

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Crawler statistics, thread-safe
type CrawlerStats struct {
	PagesCrawled      []uint32
	QueueSize         []int
	PagesSkippedErr   int
	PagesSkippedLang  int
	DuplicatesAvoided int
	LastUpdated       time.Time
	mu                sync.Mutex
	stopChan          chan struct{}

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
func (stats *CrawlerStats) StartWriter(interval time.Duration, filename string) {
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

				// write to json file
				jsonData, err := json.MarshalIndent(stats, "", "  ")
				stats.mu.Unlock()
				if err != nil {
					fmt.Println("Error serializing CrawlerStats:", err)
					continue
				}

				err = os.WriteFile(filename, jsonData, 0644)
				if err != nil {
					fmt.Println("Error writing to stats file: ", err)
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
