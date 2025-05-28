package storage

import (
	"context"
	"os"

	"github.com/Jailior/open-search/backend/internal/models"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB Database interface
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

func (db *Database) AddCollection(dbname string, collectionname string, key string) {
	db.collection[key] = db.client.Database(dbname).Collection(collectionname)
}

// Inserts a PageData page as a document in collection
func (db *Database) InsertRawPage(pd *models.PageData, key string) error {
	_, err := db.collection[key].InsertOne(db.ctx, *pd)
	db.Error = err
	return err
}
