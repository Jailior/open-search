package indexer

import (
	"context"
	"log"
	"time"

	"github.com/Jailior/open-search/backend/internal/models"
	"github.com/Jailior/open-search/backend/internal/parsing"
	"github.com/Jailior/open-search/backend/internal/storage"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
)

const DB_NAME = "opensearch"
const PAGE_INSERT_COLLECTION = "pages"
const PAGE_INDEX_COLLECTION = "inverted_index"

type Indexer struct {
	GroupName   string
	StreamName  string
	Database    *storage.Database
	RedisClient *storage.RedisClient
}

func (idx *Indexer) RunWorker(cancelContext context.Context, consumerName string) {

	log.Printf("[%s] started\n", consumerName)

	for {
		select {
		case <-cancelContext.Done():
			log.Printf("[%s] Shutdown signal received", consumerName)
			return
		default:
			messages, err := idx.RedisClient.ReadStream(idx.StreamName, idx.GroupName, consumerName)
			if err != nil {
				time.Sleep(2 * time.Second)
				continue
			}
			if messages != nil && len(messages) > 0 {
				idx.ProcessMessages(messages)
			}
		}
	}
}

// Constructs an inverted index based on page
func (idx *Indexer) IndexPage(docId string, page *models.PageData) error {
	terms := parsing.TokenizeText(page.Title + " " + page.Content)
	termsLength := float64(len(terms))

	for term, positions := range terms {
		termFreq := float64(len(positions)) / termsLength
		filter := bson.M{"term": term}

		posting := models.IndexerPosting{
			DocID:     docId,
			Title:     page.Title,
			URL:       page.URL,
			TF:        termFreq,
			Positions: positions,
		}

		update := bson.M{
			"$addToSet": bson.M{
				"postings": posting,
			},
			"$setOnInsert": bson.M{
				"DF": 0, // initialize DF
			},
		}

		// update term in database
		result, err := idx.Database.UpdateTerm(PAGE_INDEX_COLLECTION, filter, update)
		if err != nil {
			log.Println("Failed to update index for term", term, ": ", err)
			continue
		}

		if result.ModifiedCount > 0 || result.UpsertedCount > 0 {
			idx.Database.IncrementDF(PAGE_INDEX_COLLECTION, filter)
		}
	}
	idx.Database.IncrementDocCount(PAGE_INDEX_COLLECTION)
	return nil
}

// Processes the entries from a Redis stream for indexing
func (idx *Indexer) ProcessMessages(messages []redis.XMessage) {
	rd := idx.RedisClient
	db := idx.Database

	for _, message := range messages {
		// get doc _id
		idVal := message.Values["id"]
		idStr, ok := idVal.(string)
		if !ok {
			log.Println("Invalid id value in stream message")
			continue
		}

		// Get page from database
		page, err := db.FetchRawPage(idStr, PAGE_INSERT_COLLECTION)
		if err != nil {
			log.Println("Mongo fetch error: ", err)
			continue
		}

		// Index Page
		log.Println("Title: ", page.Title)
		log.Println("URL: ", page.URL)
		err = idx.IndexPage(idStr, page)

		// acknowledgement of reading message
		_, err = rd.Client.XAck(rd.Ctx, idx.StreamName, idx.GroupName, message.ID).Result()
		if err != nil {
			log.Println("Failed to ack message: ", err)
		}
	}
}
