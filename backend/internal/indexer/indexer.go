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

// Mongo database name
const DB_NAME = "opensearch"

// Raw page collection
const PAGE_INSERT_COLLECTION = "pages"

// Inverted index collection
const PAGE_INDEX_COLLECTION = "inverted_index"

// Indexer context
type Indexer struct {
	GroupName   string
	StreamName  string
	Database    *storage.Database
	RedisClient *storage.RedisClient
}

// Initializes Indexer worker, runs until shutdown received on cancel context
func (idx *Indexer) RunWorker(cancelContext context.Context, consumerName string) {

	log.Printf("[%s] started\n", consumerName)

	for {
		select {
		// selects between shutdown and default behaviour
		// shutting down
		case <-cancelContext.Done():
			log.Printf("[%s] Shutdown signal received", consumerName)
			return
		default:
			// default behaviour

			// reads set of messages from Redis stream
			messages, err := idx.RedisClient.ReadStream(idx.StreamName, idx.GroupName, consumerName)

			// if read error, wait
			if err != nil {
				time.Sleep(2 * time.Second)
				continue
			}

			// process messages if valid
			if messages != nil && len(messages) > 0 {
				idx.ProcessMessages(messages)
			}
		}
	}
}

// Constructs an inverted index based on a page
func (idx *Indexer) IndexPage(docId string, page *models.PageData) error {
	// get terms in page and their positions in the text
	terms := parsing.TokenizeText(page.Title + " " + page.Content)
	termsLength := float64(len(terms))

	// for each term get TF and add page as a posting
	for term, positions := range terms {
		// get term frequency
		termFreq := float64(len(positions)) / termsLength

		// filter based on term
		filter := bson.M{"term": term}

		// posting stored in database
		posting := models.IndexerPosting{
			DocID:     docId,
			Title:     page.Title,
			URL:       page.URL,
			TF:        termFreq,
			Positions: positions,
		}

		// Update Option: update term document with posting or add it if it doesn't exit
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

		// Increment document frequency if posting sucessfully added
		if result.ModifiedCount > 0 || result.UpsertedCount > 0 {
			idx.Database.IncrementDF(PAGE_INDEX_COLLECTION, filter)
		}
	}
	// Increment the number of total pages referred to by the inverted index
	idx.Database.IncrementDocCount(PAGE_INDEX_COLLECTION)
	return nil
}

// Processes the entries from a Redis stream for indexing
func (idx *Indexer) ProcessMessages(messages []redis.XMessage) {
	rd := idx.RedisClient
	db := idx.Database

	// for each message, id, get page the id refers to from database and index it
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

		// Acknowledge indexing page on shared Redis stream
		_, err = rd.Client.XAck(rd.Ctx, idx.StreamName, idx.GroupName, message.ID).Result()
		if err != nil {
			log.Println("FAILED to ACK message: ", err)
		}

	}
}
