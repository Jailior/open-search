package indexer

import (
	"log"
	"strings"
	"unicode"

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

// returns words and their positions in the text
func Tokenize(text string) map[string][]int {
	words := make(map[string][]int)

	// split by whitespace, lowercase and strip puncuation
	tokens := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	for pos, raw := range tokens {
		word := strings.ToLower(raw)
		if parsing.Stopwords[word] || word == "" {
			continue
		}
		words[word] = append(words[word], pos)
	}
	return words
}

// Constructs an inverted index based on page
func (idx *Indexer) IndexPage(docId string, page *models.PageData) error {
	terms := Tokenize(page.Title + " " + page.Content)
	termsLength := float64(len(terms))

	for term, positions := range terms {
		termFreq := float64(len(positions)) / termsLength
		filter := bson.M{"term": term}

		posting := bson.M{
			"doc_id":    docId,
			"url":       page.URL,
			"TF":        termFreq,
			"positions": positions,
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
	return nil
}

// Processes the entries from a Redis stream for indexing
func (idx *Indexer) ProcessEntries(entries []redis.XStream) {
	rd := idx.RedisClient
	db := idx.Database

	for _, stream := range entries {
		for _, message := range stream.Messages {
			// get doc _id
			idVal := message.Values["id"]
			idStr, ok := idVal.(string)
			if !ok {
				log.Println("Invalid id value in stream message")
				continue
			}

			// Get page from database
			page, err := db.FetchPage(idStr, PAGE_INSERT_COLLECTION)
			if err != nil {
				log.Println("Mongo fetch error: ", err)
				continue
			}

			// Index Page
			log.Println("Indexing page: ", page.Title)
			err = idx.IndexPage(idStr, page)

			// acknowledgement of reading
			_, err = rd.Client.XAck(rd.Ctx, idx.StreamName, idx.GroupName, message.ID).Result()
			if err != nil {
				log.Println("Failed to ack message: ", err)
			}
		}
	}
}
