package pagerank

import (
	"context"
	"fmt"
	"log"

	"github.com/Jailior/open-search/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Graph structure, uses adjacency list representation
type Graph struct {
	vertices map[string][]string // adjacency list used for outlinks
}

// Makes empty graph
func MakeGraph() *Graph {
	return &Graph{
		vertices: make(map[string][]string),
	}
}

// Adds vertex to graph, returns true if non-duplicate vertex is successfully added
func (g *Graph) AddVertex(url string) bool {
	if _, ok := g.vertices[url]; !ok {
		g.vertices[url] = make([]string, 0)
		return true
	}
	return false
}

// Adds edge between two vertices, returns false if either vertex is not present in graph
func (g *Graph) AddEdge(url1, url2 string) bool {
	_, ok1 := g.vertices[url1]
	_, ok2 := g.vertices[url2]
	if !ok1 || !ok2 {
		return false
	}
	g.vertices[url1] = append(g.vertices[url1], url2)
	return true
}

// Builds up graph from given collection, expects url and outlinks
func (g *Graph) BuildFromPages(collection *mongo.Collection, ctx context.Context) error {
	// retrieve all crawled pages' URLs
	validPages := make(map[string]struct{})
	cursor, err := collection.Find(ctx, map[string]interface{}{}, options.Find().SetProjection(bson.M{"url": 1}))
	if err != nil {
		return fmt.Errorf("Failed to read collection pages: %w", err)
	}
	// for all page URLs found decode and add to unique page map
	for cursor.Next(ctx) {
		var page struct {
			URL string
		}
		cursor.Decode(&page)
		validPages[page.URL] = struct{}{}
	}
	cursor.Close(ctx)

	// retrieve whole page collection
	cursor, err = collection.Find(ctx, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("Failed to read collection pages: %w", err)
	}
	defer cursor.Close(ctx)

	var totalPages, totalEdges, crawledPages int

	// building graph from pages
	for cursor.Next(ctx) {
		// get page model instance
		var page models.PageData
		err := cursor.Decode(&page)
		if err != nil {
			log.Println("Error decoding page: ", err)
			continue
		}

		// add page
		g.AddVertex(page.URL)
		crawledPages++

		// add all outlinks
		for _, outlink := range page.Outlinks {
			_, exists := validPages[outlink]
			if exists {
				g.AddVertex(outlink)
				g.AddEdge(page.URL, outlink)
				totalEdges++
			}
		}
	}

	totalPages = len(g.vertices)
	log.Printf("Built graph with %d crawled pages, %d total pages and %d edges\n", crawledPages, totalPages, totalEdges)

	return nil
}
