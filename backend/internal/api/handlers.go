package api

import (
	"net/http"

	"github.com/Jailior/open-search/backend/internal/parsing"
	"github.com/Jailior/open-search/backend/internal/storage"
	"github.com/gin-gonic/gin"
)

// Database Wrapper
type SearchService struct {
	DB        *storage.Database
	COLL_NAME string
}

type DocScore struct {
	DocID string
	Title string
	URL   string
	Score float64
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (svc *SearchService) SearchHandler(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing query"})
		return
	}

	terms := parsing.TokenizeQuery(query)
	if len(terms) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No valid terms found"})
		return
	}

	// fetch total doc count
	svc.DB.TotalDocCount(svc.COLL_NAME)

	// scores := map[string]*DocScore{}

	// get posting for each term
	// for _, term := terms {

	// }

	c.JSON(http.StatusOK, gin.H{
		"query":   query,
		"results": "Insert search results here",
	})

}

func MetricsHandler(c *gin.Context) {

	// TODO: read updated metrics here

	c.JSON(http.StatusOK, gin.H{"metrics": "put metrics here"})
}
