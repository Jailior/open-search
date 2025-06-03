package api

import (
	"context"
	"log"
	"math"
	"net/http"
	"sort"

	"github.com/Jailior/open-search/backend/internal/parsing"
	"github.com/Jailior/open-search/backend/internal/storage"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// Database Wrapper
type SearchService struct {
	DB *storage.Database
}

// Returned struct by API, representing a page
type DocScore struct {
	DocID   string  `json:"doc_id"`
	Title   string  `json:"title"`
	URL     string  `json:"url"`
	Snippet string  `json:"snippet"`
	Score   float64 `json:"score"`
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

	// pagination parameters
	limit := getLimitQuery(c)
	offset := getOffsetQuery(c)

	// fetch total doc count
	docCount, _ := svc.DB.TotalDocCount(COLL_NAME)

	scores := map[string]*DocScore{}

	// get posting for each term
	for _, term := range terms {
		entry, err := svc.DB.FetchPostings(term, COLL_NAME)
		if err != nil {
			continue
		}

		idf := math.Log(float64(docCount) / float64(entry.DF))

		for _, posting := range entry.Postings {
			tfIdf := posting.TF * idf
			if _, ok := scores[posting.DocID]; !ok {
				page, _ := svc.DB.FetchRawPage(posting.DocID, "pages")
				snippet := getSnippet(posting.Positions, page.Content, term)
				scores[posting.DocID] = &DocScore{
					DocID:   posting.DocID,
					Title:   posting.Title,
					URL:     posting.URL,
					Snippet: snippet,
				}
			}
			scores[posting.DocID].Score += tfIdf
		}
	}

	ranked := make([]*DocScore, 0, len(scores))

	for _, doc := range scores {
		ranked = append(ranked, doc)
	}

	// sort by score
	sort.Slice(ranked, func(i int, j int) bool {
		return ranked[i].Score > ranked[j].Score
	})

	// pagination slicing
	numPages := len(ranked)
	start := min(offset, numPages)
	end := min(start+limit, numPages)
	paged := ranked[start:end]

	// update metrics
	err := svc.IncrementSearchNum()
	if err != nil {
		log.Println("Error incrementing search number: ", err)
	}

	// return results and query
	c.JSON(200, gin.H{
		"query":        query,
		"totalResults": numPages,
		"results":      paged,
	})

}

func (svc *SearchService) MetricsHandler(c *gin.Context) {

	var crawlerStats struct {
		PagesCrawled      []uint32 `bson:"pages_crawled" json:"pages_crawled"`
		QueueSize         []int    `bson:"queue_size" json:"queue_size"`
		PagesSkippedErr   int      `bson:"page_errs" json:"page_errs"`
		PagesSkippedLang  int      `bson:"pages_skipped_lang" json:"pages_skipped_lang"`
		DuplicatesAvoided int      `bson:"duplicates_avoided" json:"duplicates_avoided"`
		NumberOfSearchs   int      `bson:"number_of_searches" json:"number_of_searches"`
	}

	err := svc.DB.GetCollection("pages").FindOne(context.Background(), bson.M{"_id": "crawler_stats"}).Decode(&crawlerStats)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "metrics yet to be initialized, or not found."})
	}

	c.JSON(http.StatusOK, gin.H{"metrics": crawlerStats})
}
