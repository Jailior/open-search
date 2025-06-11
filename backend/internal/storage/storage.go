package storage

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Jailior/open-search/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB Database Interface
// Currently supports only on DB Name and Collection
type Database struct {
	client     *mongo.Client
	collection map[string]*mongo.Collection
	ctx        context.Context
	Error      error
}

// Returns default Database, only ctx is initialized
func MakeDB() *Database {
	return &Database{
		client:     nil,
		collection: make(map[string]*mongo.Collection),
		ctx:        context.TODO(),
		Error:      nil,
	}
}

// Connects to MongoDB database at env variable MONGODB_URI
func (db *Database) Connect() {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		panic("MONGODB_URI environment variable is not set")
	}
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	var err error
	db.client, err = mongo.Connect(db.ctx, opts)

	if err != nil {
		panic(err)
	}
}

// Disconnects from database
func (db *Database) Disconnect() {
	if err := db.client.Disconnect(db.ctx); err != nil {
		panic(err)
	}
}

// Getter, returns database context
func (db *Database) GetContext() *context.Context {
	return &db.ctx
}

// Adds collection with collection name to database, must be done before any functions called using this collection
func (db *Database) AddCollection(dbname string, collectionname string) {
	db.collection[collectionname] = db.client.Database(dbname).Collection(collectionname)
}

// Returns a reference to a collection stored in Database
func (db *Database) GetCollection(collectionname string) *mongo.Collection {
	return db.collection[collectionname]
}

// Makes an index on field, enforces uniqueness
func (db *Database) MakeIndex(collectionname, field string) error {
	indexModel := mongo.IndexModel{
		Keys: bson.M{
			field: 1,
		},
		Options: options.Index().SetUnique(true),
	}
	_, err := db.GetCollection(collectionname).Indexes().CreateOne(db.ctx, indexModel)
	return err
}

// Inserts a PageData page as a document in collection
func (db *Database) InsertRawPage(pd *models.PageData, collectionname string) (string, error) {
	res, err := db.GetCollection(collectionname).InsertOne(db.ctx, *pd)
	db.Error = err
	if err != nil {
		if we, ok := err.(mongo.WriteException); ok {
			for _, e := range we.WriteErrors {
				if e.Code == 11000 {
					// duplicate key error
					return "", fmt.Errorf("DUPE: Page Error: %w", err)
				}
			}
		}
		db.Error = err
		return "", err
	}
	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

// Fetches a page with doc _id idHex from the collection
func (db *Database) FetchRawPage(idHex string, collectionname string) (*models.PageData, error) {
	id, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return nil, err
	}

	var result models.PageData
	err = db.collection[collectionname].FindOne(db.ctx, bson.M{"_id": id}).Decode(&result)
	return &result, err
}

// Batch fetches from raw page collection form a list of ObjectIDs
func (db *Database) FetchRawPageBatch(idHexes []string, collectionname string) ([]models.PageData, error) {
	var objIDs []primitive.ObjectID
	for _, hex := range idHexes {
		id, err := primitive.ObjectIDFromHex(hex)
		if err != nil {
			continue
		}
		objIDs = append(objIDs, id)
	}

	cursor, err := db.collection[collectionname].Find(db.ctx, bson.M{
		"_id": bson.M{"$in": objIDs},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(db.ctx)

	var results []models.PageData
	for cursor.Next(db.ctx) {
		var page models.PageData
		if err := cursor.Decode(&page); err != nil {
			continue
		}
		results = append(results, page)
	}

	return results, nil
}

// Gets postings for a term
func (db *Database) FetchPostings(term string, collectionname string) (*models.TermEntry, error) {
	var result models.TermEntry
	err := db.collection[collectionname].FindOne(db.ctx, bson.M{"term": term}).Decode(&result)
	if err != nil || result.DF == 0 {
		return nil, err
	}
	return &result, nil
}

// Batch fetches all postings for a list of terms
func (db *Database) FetchPostingsBatch(terms []string, collectionname string) ([]models.TermEntry, error) {
	cursor, err := db.collection[collectionname].Find(db.ctx, bson.M{
		"term": bson.M{"$in": terms},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(db.ctx)

	var results []models.TermEntry
	for cursor.Next(db.ctx) {
		var entry models.TermEntry
		if err := cursor.Decode(&entry); err != nil {
			continue
		}
		if entry.DF > 0 {
			results = append(results, entry)
		}
	}
	return results, nil
}

// Updates a term in the collection based on a filter and update bson.M
// Returns a reference to the mongo UpdateResult from the update
func (db *Database) UpdateTerm(collectionname string, filter bson.M, update bson.M) (*mongo.UpdateResult, error) {
	collection := db.GetCollection(collectionname)
	if collection == nil {
		log.Println("Collection with name", collectionname, "not in Database struct")
		db.Error = fmt.Errorf("Collection name not found.")
		return nil, db.Error
	}
	result, err := collection.UpdateOne(db.ctx, filter, update, options.Update().SetUpsert(true))
	return result, err
}

// Increments the DF key's value in a document selected via filter
// Assumes the document has a DF key
func (db *Database) IncrementDF(collectionname string, filter bson.M) error {
	collection := db.GetCollection(collectionname)
	if collection == nil {
		log.Println("Collection with name", collectionname, "not in Database struct")
		db.Error = fmt.Errorf("Collection name not found.")
		return db.Error
	}
	_, _ = collection.UpdateOne(
		db.ctx,
		filter,
		bson.M{"$inc": bson.M{"DF": 1}},
	)
	return nil
}

// Initalizes a corpus stats document in collection
func (db *Database) InitializeIndexCorpus(collectionname string) error {
	meta := bson.M{
		"_id":         "corpus_stats",
		"total_pages": 0,
		"term":        "",
	}
	_, db.Error = db.collection[collectionname].InsertOne(db.ctx, meta)
	return db.Error
}

// Returns the total_pages in corpus stats
func (db *Database) TotalDocCount(collectionname string) (int, error) {
	var result struct {
		TotalDocCount int `bson:"total_pages"`
	}
	err := db.collection[collectionname].FindOne(db.ctx, bson.M{"_id": "corpus_stats"}).Decode(&result)
	return result.TotalDocCount, err
}

// Increments the total_pages in corpus stats
func (db *Database) IncrementDocCount(collectionname string) error {
	_, err := db.collection[collectionname].UpdateByID(db.ctx, "corpus_stats", bson.M{"$inc": bson.M{"total_pages": 1}})
	return err
}

// Gets a pagerank score for a url
func (db *Database) GetPageRank(url string) float64 {
	collection := db.GetCollection("pagerank")
	filter := bson.M{"url": url}
	var result struct {
		Score float64 `bson:"score"`
	}
	err := collection.FindOne(db.ctx, filter).Decode(&result)
	if err != nil {
		return 0.0
	}
	return result.Score
}

// Batch fetches pagerank scores for a list of urls
func (db *Database) FetchPageRankBatch(urls []string) (map[string]float64, error) {
	cursor, err := db.collection["pagerank"].Find(db.ctx, bson.M{
		"url": bson.M{"$in": urls},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(db.ctx)

	results := make(map[string]float64)
	for cursor.Next(db.ctx) {
		var doc struct {
			URL   string  `bson:"url"`
			Score float64 `bson:"score"`
		}
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		results[doc.URL] = doc.Score
	}

	return results, nil
}
