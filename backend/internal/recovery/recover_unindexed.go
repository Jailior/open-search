package main

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Recovers Redis pages to index stream due to premature ACK error

func main() {
	ctx := context.Background()

	mongoURI := "mongodb://localhost:27017"
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Mongo connection failed:", err)
	}
	defer client.Disconnect(ctx)

	db := client.Database("opensearch")
	pagesColl := db.Collection("pages")
	indexColl := db.Collection("inverted_index")

	// Redis setup
	redisAddr := "localhost:6379"
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0,
	})

	streamName := "pages_to_index"

	// 1. Get all _id from pages collection
	pageIDs := make(map[string]struct{})
	cursor, err := pagesColl.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"_id": 1}))
	if err != nil {
		log.Fatal("Failed to fetch pages:", err)
	}
	for cursor.Next(ctx) {
		var doc struct {
			ID primitive.ObjectID `bson:"_id"`
		}
		if err := cursor.Decode(&doc); err == nil {
			pageIDs[doc.ID.Hex()] = struct{}{}
		}
	}
	cursor.Close(ctx)

	// 2. Get all doc_ids from inverted index
	indexedIDs := make(map[string]struct{})
	cursor, err = indexColl.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"postings.doc_id": 1}))
	if err != nil {
		log.Fatal("Failed to fetch indexed documents:", err)
	}
	for cursor.Next(ctx) {
		var entry struct {
			Postings []struct {
				DocID string `bson:"doc_id"`
			} `bson:"postings"`
		}
		if err := cursor.Decode(&entry); err == nil {
			for _, p := range entry.Postings {
				indexedIDs[p.DocID] = struct{}{}
			}
		}
	}
	cursor.Close(ctx)

	// 3. Find missing IDs
	var missing []string
	for id := range pageIDs {
		if _, found := indexedIDs[id]; !found {
			missing = append(missing, id)
		}
	}

	// 4. Push missing IDs to Redis stream
	for _, id := range missing {
		err := rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: streamName,
			Values: map[string]interface{}{"id": id},
		}).Err()
		if err != nil {
			log.Printf("⚠️ Failed to push %s to stream: %v", id, err)
		}
	}

	fmt.Printf("✅ Requeued %d unindexed pages\n", len(missing))
}
