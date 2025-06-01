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

func getSnippet(positions []int, text string) string {
	if len(text) == 0 || len(positions) == 0 {
		return ""
	}

	words := parsing.SplitWords(text)
	pos := positions[0]

	start := max(0, pos-10)
	end := min(len(words), pos+10)

	snippetWords := words[start:end]
	snippet := strings.Join(snippetWords, " ")

	if len(snippet) > 0 {
		snippet = strings.ToUpper(snippet[:1]) + snippet[1:]
	}
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(words) {
		snippet = snippet + "..."
	}

	return snippet
}

func getLimitQuery(c *gin.Context) int {
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			return parsed
		}
	}
	return DEFAULT_PAGE_LIMIT
}

func getOffsetQuery(c *gin.Context) int {
	if off := c.Query("offset"); off != "" {
		if parsed, err := strconv.Atoi(off); err == nil && parsed >= 0 {
			return parsed
		}
	}
	return DEFAULT_PAGE_OFFSET
}

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

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
