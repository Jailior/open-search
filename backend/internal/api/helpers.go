package api

import (
	"context"
	"strconv"
	"strings"

	"github.com/Jailior/open-search/backend/internal/parsing"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/* Helper Functions */

// Gets a snippet from a page relevant to the term,
// assumes positions points to the positions in the text where the term is present
func getSnippet(positions []int, text string, term string) string {
	if len(text) == 0 || len(positions) == 0 {
		return ""
	}

	// split words into list
	words := parsing.SplitWords(text)
	// get first position of term
	pos := positions[0]

	// start and end positions of snippet
	start := max(0, pos-10)
	end := min(len(words), pos+10)

	// limit to start and end
	snippetWords := words[start:end]

	// highlight/bold term in snippet
	for i, word := range snippetWords {
		if strings.EqualFold(word, term) {
			snippetWords[i] = "<strong>" + word + "</strong>"
		}
	}

	// join snippet back into string
	snippet := strings.Join(snippetWords, " ")

	// check if valid snippet
	if len(snippet) > 0 {
		snippet = strings.ToUpper(snippet[:1]) + snippet[1:]
	}

	// add elipsis
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(words) {
		snippet = snippet + "..."
	}

	return snippet
}

// Extracts the limit query from a search request
// Returns DEFAULT_PAGE_LIMIT if not specified
func getLimitQuery(c *gin.Context) int {
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			return parsed
		}
	}
	return DEFAULT_PAGE_LIMIT
}

// Extracts the offset query from a search request
// Returns DEFAULT_PAGE_LIMIT if not specified
func getOffsetQuery(c *gin.Context) int {
	if off := c.Query("offset"); off != "" {
		if parsed, err := strconv.Atoi(off); err == nil && parsed >= 0 {
			return parsed
		}
	}
	return DEFAULT_PAGE_OFFSET
}

// Increments the number_of_searches field in the crawler stats document
func (svc *SearchService) IncrementSearchNum() error {
	collection := svc.DB.GetCollection("pages")

	filter := bson.M{"_id": "crawler_stats"}

	update := bson.M{
		"$inc": bson.M{"number_of_searches": 1},
	}

	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(context.Background(), filter, update, opts)
	return err
}

// Returns the min of a and b
func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

// Returns the max of a and b
func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
