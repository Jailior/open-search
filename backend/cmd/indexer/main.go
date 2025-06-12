package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Jailior/open-search/backend/internal/indexer"
	"github.com/Jailior/open-search/backend/internal/storage"
)

// Redis stream name for pages to be indexed
const REDIS_INDEX_QUEUE = "pages_to_index"

// Redis stream consumer group
const REDIS_STREAM_GROUP = "indexer_group"

func main() {

	// initialize flags
	workers := flag.Int("workers", 8, "Number of concurrent indexer workers")
	reset := flag.Bool("reset", false, "Clear Redis indexer stream before indexing.")

	flag.Parse()

	// make and connect Redis and Mongo clients
	rd := storage.MakeRedisClient()
	db := storage.MakeDB()
	db.Connect()
	defer db.Disconnect()

	// Add raw page, inverted index collections
	db.AddCollection(indexer.DB_NAME, indexer.PAGE_INSERT_COLLECTION)
	db.AddCollection(indexer.DB_NAME, indexer.PAGE_INDEX_COLLECTION)

	// add corpus stats document in inverted index
	db.InitializeIndexCorpus(indexer.PAGE_INDEX_COLLECTION)

	// if reset flag was passed, reset Redis stream and initialize index on "term" field
	if *reset {
		rd.ResetStream(REDIS_INDEX_QUEUE)
		// db.MakeIndex(indexer.PAGE_INDEX_COLLECTION, "term")
	}

	// create consumer if not already created
	rd.EnsureConsumerGroup(REDIS_INDEX_QUEUE, REDIS_STREAM_GROUP)

	// shutdown context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// capture interrupt or term signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// send signal on sigs channel to indexers when signal received
	go func() {
		<-sigs
		log.Println("Indexer shutdown signal received...")
		cancel()
	}()

	// initialize indexers
	var wg sync.WaitGroup
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			consumerName := fmt.Sprintf("indexer-%d", workerID)
			idx := &indexer.Indexer{
				Database:    db,
				RedisClient: rd,
				StreamName:  REDIS_INDEX_QUEUE,
				GroupName:   REDIS_STREAM_GROUP,
			}
			idx.RunWorker(ctx, consumerName)
		}(i)
	}

	wg.Wait()
	log.Println("All indexer workers shut down.")

}
