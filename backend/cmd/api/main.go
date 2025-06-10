package main

import (
	"log"

	"github.com/Jailior/open-search/backend/internal/api"
	"github.com/Jailior/open-search/backend/internal/storage"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	db := storage.MakeDB()
	db.Connect()
	defer db.Disconnect()
	db.AddCollection(api.DB_NAME, api.COLL_NAME)
	db.AddCollection(api.DB_NAME, "pages")
	db.AddCollection(api.DB_NAME, "pagerank")

	db.InitializeIndexCorpus(api.COLL_NAME)

	svc := &api.SearchService{DB: db}

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://opensearchengine.app"},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		AllowCredentials: true,
	}))

	api.SetUpRouter(router, svc)

	log.Println("Starting API on localhost:8080")
	err := router.Run(":8080")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
