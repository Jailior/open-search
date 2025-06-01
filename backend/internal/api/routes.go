package api

import (
	"github.com/gin-gonic/gin"
)

const DB_NAME = "opensearch"
const COLL_NAME = "inverted_index"

const DEFAULT_PAGE_LIMIT = 10
const DEFAULT_PAGE_OFFSET = 0

func SetUpRouter(router *gin.Engine, svc *SearchService) {
	router.GET("/health", HealthCheck)
	router.GET("/metrics", svc.MetricsHandler)
	router.GET("/search", svc.SearchHandler)
}
