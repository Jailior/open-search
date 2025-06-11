package api

import (
	"github.com/gin-gonic/gin"
)

// Database name and collection name used by api
const DB_NAME = "opensearch"
const COLL_NAME = "inverted_index"

// Default page limit and offset for pagination if not specified
const DEFAULT_PAGE_LIMIT = 10
const DEFAULT_PAGE_OFFSET = 0

// Sets up handlers for health, metrics and search
func SetUpRouter(router *gin.Engine, svc *SearchService) {
	router.GET("/health", HealthCheck)
	router.GET("/metrics", svc.MetricsHandler)
	router.GET("/search", svc.SearchHandler)
}
