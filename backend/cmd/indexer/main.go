package main

import (
	"time"

	"github.com/Jailior/open-search/backend/internal/indexer"
	"github.com/Jailior/open-search/backend/internal/storage"
)

const REDIS_INDEX_QUEUE = "pages_to_index"
const REDIS_STREAM_GROUP = "indexer_group"
const REDIS_CONSUMER_NAME = "indexer-1"

func main() {

	rd := storage.MakeRedisClient()
	db := storage.MakeDB()
	db.Connect()
	defer db.Disconnect()

	db.AddCollection(indexer.DB_NAME, indexer.PAGE_INSERT_COLLECTION)
	db.AddCollection(indexer.DB_NAME, indexer.PAGE_INDEX_COLLECTION)
	db.InitializeIndexCorpus(indexer.PAGE_INDEX_COLLECTION)

	idx := &indexer.Indexer{
		Database:    db,
		RedisClient: rd,
		StreamName:  REDIS_INDEX_QUEUE,
		GroupName:   REDIS_STREAM_GROUP,
	}

	for {
		messages, err := rd.ReadStream(idx.StreamName, idx.GroupName, REDIS_CONSUMER_NAME)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		idx.ProcessEntries(messages)
	}
}
