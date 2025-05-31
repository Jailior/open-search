package main

import (
	"log"

	"github.com/Jailior/open-search/backend/internal/api"
	"github.com/Jailior/open-search/backend/internal/storage"
	"github.com/gin-gonic/gin"
)

const DB_NAME = "opensearch"
const COLL_NAME = "inverted_index"

func main() {

	db := storage.MakeDB()
	db.Connect()
	defer db.Disconnect()
	db.AddCollection(DB_NAME, COLL_NAME)
	log.Println("Collection added.")

	db.InitializeIndexCorpus(COLL_NAME)

	svc := &api.SearchService{DB: db, COLL_NAME: COLL_NAME}

	router := gin.Default()

	api.SetUpRouter(router, svc)

	log.Println("Starting API on localhost:8080")
	err := router.Run(":8080")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
