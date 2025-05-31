package api

import (
	"github.com/gin-gonic/gin"
)

func SetUpRouter(router *gin.Engine, svc *SearchService) {
	router.GET("/health", HealthCheck)
	router.GET("/metrics", MetricsHandler)
	router.GET("/search", svc.SearchHandler)
}
