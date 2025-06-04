package pagerank

import (
	"context"
	"fmt"
	"log"

	"github.com/Jailior/open-search/backend/internal/models"
	"go.mongodb.org/mongo-driver/mongo"
)

type Graph struct {
	vertices map[string][]string // adjacency list for outlinks
}

func MakeGraph() *Graph {
	return &Graph{
		vertices: make(map[string][]string),
	}
}

func (g *Graph) AddVertex(url string) bool {
	if _, ok := g.vertices[url]; !ok {
		g.vertices[url] = make([]string, 0)
		return true
	}
	return false
}

func (g *Graph) AddEdge(url1, url2 string) bool {
	_, ok1 := g.vertices[url1]
	_, ok2 := g.vertices[url2]
	if !ok1 || !ok2 {
		return false
	}
	g.vertices[url1] = append(g.vertices[url1], url2)
	return true
}

func (g *Graph) BuildFromPages(collection *mongo.Collection, ctx context.Context) error {
	// retrieve whole collection
	cursor, err := collection.Find(ctx, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("Failed to read collection pages: %w", err)
	}
	defer cursor.Close(ctx)

	var totalPages, totalEdges, crawledPages int

	for cursor.Next(ctx) {
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
			g.AddVertex(outlink)

			ok := g.AddEdge(page.URL, outlink)
			if ok {
				totalEdges++
			}
		}
	}

	totalPages = len(g.vertices)
	log.Printf("Built graph with %d crawled pages, %d total pages and %d edges\n", crawledPages, totalPages, totalEdges)

	return nil
}
