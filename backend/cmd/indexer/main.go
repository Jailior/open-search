package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Jailior/open-search/backend/internal/indexer"
	"github.com/Jailior/open-search/backend/internal/storage"
)

const REDIS_INDEX_QUEUE = "pages_to_index"
const REDIS_STREAM_GROUP = "indexer_group"
const REDIS_CONSUMER_NAME = "indexer-1"

const NUM_WORKERS = 8

func main() {

	rd := storage.MakeRedisClient()

	db := storage.MakeDB()
	db.Connect()
	defer db.Disconnect()

	db.AddCollection(indexer.DB_NAME, indexer.PAGE_INSERT_COLLECTION)
	db.AddCollection(indexer.DB_NAME, indexer.PAGE_INDEX_COLLECTION)
	db.InitializeIndexCorpus(indexer.PAGE_INDEX_COLLECTION)

	// create consumer if not already created
	rd.EnsureConsumerGroup(REDIS_INDEX_QUEUE, REDIS_STREAM_GROUP)

	// shutdown context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// capture interrupt or term signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		log.Println("Indexer shutdown signal received...")
		cancel()
	}()

	var wg sync.WaitGroup
	for i := 0; i < NUM_WORKERS; i++ {
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
