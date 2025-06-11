package main

import (
	"log"

	"github.com/Jailior/open-search/backend/internal/api"
	"github.com/Jailior/open-search/backend/internal/storage"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

/*
Main function to run API handlers
*/
func main() {

	// Set up database and connect both raw pages, pagerank and index collections
	db := storage.MakeDB()
	db.Connect()
	defer db.Disconnect()
	db.AddCollection(api.DB_NAME, api.COLL_NAME)
	db.AddCollection(api.DB_NAME, "pages")
	db.AddCollection(api.DB_NAME, "pagerank")

	// initialize corpus stats if not already initialized
	db.InitializeIndexCorpus(api.COLL_NAME)

	// give SearchService wrapper access to database reference
	svc := &api.SearchService{DB: db}

	router := gin.Default()

	// CORS middleware configuration allowing requests from frontend only
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://opensearchengine.app"},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		AllowCredentials: true,
	}))

	// set up router with handler endpoints
	api.SetUpRouter(router, svc)

	// log and run on port 8080 with error handling
	log.Println("Starting API on localhost:8080")
	err := router.Run(":8080")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
