package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func SearchHandler(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing query"})
		return
	}

	// TODO: add TF-IDF scoring and results here

	c.JSON(http.StatusOK, gin.H{
		"query":   query,
		"results": "Insert search results here",
	})

}

func MetricsHandler(c *gin.Context) {

	// TODO: read updated metrics here

	c.JSON(http.StatusOK, gin.H{"metrics": "put metrics here"})
}
