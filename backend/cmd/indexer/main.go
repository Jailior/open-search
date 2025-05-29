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
	db.AddCollection(indexer.DB_NAME, indexer.PAGE_INSERT_COLLECTION)
	db.AddCollection(indexer.DB_NAME, indexer.PAGE_INDEX_COLLECTION)
	defer db.Disconnect()

	idx := &indexer.Indexer{
		Database:    db,
		RedisClient: rd,
		StreamName:  REDIS_INDEX_QUEUE,
		GroupName:   REDIS_STREAM_GROUP,
	}

	for {
		entries, err := rd.ReadAckStream(idx.StreamName, idx.GroupName, REDIS_CONSUMER_NAME)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		idx.ProcessEntries(entries)
	}
}
