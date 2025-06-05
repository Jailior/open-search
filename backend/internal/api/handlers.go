package api

import (
	"context"
	"log"
	"math"
	"net/http"
	"sort"
	"time"

	"github.com/Jailior/open-search/backend/internal/models"
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
	DocID     string  `json:"doc_id"`
	Title     string  `json:"title"`
	URL       string  `json:"url"`
	Snippet   string  `json:"snippet"`
	Score     float64 `json:"score"`
	Positions []int   `json:"-"`
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (svc *SearchService) SearchHandler(c *gin.Context) {

	const alpha = 0.2               // tunable weight: 0.8 favors relevance (tf-IDF) and 0.2 favor authority (Page Rank Score)
	const pageRankMultiplier = 10.0 // a weight applied to pageRankScore so that both tfIdf and pagerank are ~0.1

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

	timeStart := time.Now()
	entries, err := svc.DB.FetchPostingsBatch(terms, COLL_NAME)
	if err != nil {
		log.Println("Batch fetch error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	timeAfterPostingFetch := time.Now()

	allURLs := make(map[string]struct{})
	for _, entry := range entries {
		for _, posting := range entry.Postings {
			allURLs[posting.URL] = struct{}{}
		}
	}
	uniqueURLs := make([]string, 0, len(allURLs))
	for url := range allURLs {
		uniqueURLs = append(uniqueURLs, url)
	}
	pageRankCache, _ := svc.DB.FetchPageRankBatch(uniqueURLs)

	// process entries
	for _, entry := range entries {
		idf := math.Log(float64(docCount) / float64(entry.DF))

		for _, posting := range entry.Postings {
			tfIdf := posting.TF * idf
			rank, _ := pageRankCache[posting.URL]

			weightedScore := alpha*tfIdf + (1-alpha)*rank

			if _, ok := scores[posting.DocID]; !ok {
				scores[posting.DocID] = &DocScore{
					DocID:     posting.DocID,
					Positions: posting.Positions,
				}
			}
			scores[posting.DocID].Score += weightedScore
		}
	}
	timeAfterEntries := time.Now()

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

	// docIDs list
	var docIDs []string = make([]string, 0)
	for _, page := range paged {
		docIDs = append(docIDs, page.DocID)
	}

	timeAfterMisc := time.Now()

	// batch fetch raw pages
	rawPages, err := svc.DB.FetchRawPageBatch(docIDs, "pages")

	timeAfterRawFetch := time.Now()

	// map raw page by id
	rawPageMap := make(map[string]models.PageData)
	for _, page := range rawPages {
		idHex := page.ID.Hex()
		rawPageMap[idHex] = page
	}

	// get snippets
	for _, page := range paged {
		rawPage, ok := rawPageMap[page.DocID]
		if !ok {
			continue
		}

		page.Snippet = getSnippet(page.Positions, rawPage.Content, terms[0])
		page.Title = rawPage.Title
		page.URL = rawPage.URL
	}

	timeAfterSnippets := time.Now()

	log.Println("Time taken for fetch postings batch: ", timeAfterPostingFetch.Sub(timeStart))
	log.Println("Time taken for entries processing: ", timeAfterEntries.Sub(timeAfterPostingFetch))
	log.Println("Time taken for misc ops: ", timeAfterMisc.Sub(timeAfterEntries))
	log.Println("Time taken for fetching raw pages: ", timeAfterRawFetch.Sub(timeAfterMisc))
	log.Println("Time taken for processing snippts: ", timeAfterSnippets.Sub(timeAfterRawFetch))
	log.Println("Total Time: ", timeAfterSnippets.Sub(timeStart))

	// update metrics
	err = svc.IncrementSearchNum()
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
