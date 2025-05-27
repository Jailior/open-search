package storage

import (
	"context"
	"os"

	"github.com/Jailior/open-search/backend/internal/models"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const DB_NAME = "opensearch"
const COLL_NAME = "pages"

type Database struct {
	client     *mongo.Client
	collection *mongo.Collection
	ctx        context.Context
	Error      error
}

func MakeDB() *Database {
	return &Database{
		client:     nil,
		collection: nil,
		ctx:        context.TODO(),
		Error:      nil,
	}
}

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

	db.collection = db.client.Database(DB_NAME).Collection(COLL_NAME)
}

func (db *Database) Disconnect() {
	if err := db.client.Disconnect(db.ctx); err != nil {
		panic(err)
	}
}

func (db *Database) InsertPage(pd *models.PageData) error {
	_, err := db.collection.InsertOne(db.ctx, *pd)
	db.Error = err
	return err
}
