package pagerank

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func PageRank(graph *Graph, damping float64, iterations int) map[string]float64 {
	N := len(graph.vertices)
	rank := make(map[string]float64)
	newRank := make(map[string]float64)

	// Initialize rank with all 1/N
	for url := range graph.vertices {
		rank[url] = 1.0 / float64(N)
	}

	for i := 0; i < iterations; i++ {

		// reset newRank
		for url := range graph.vertices {
			rank[url] = (1.0 - damping) / float64(N)
		}

		// rank distribution
		for url, outlinks := range graph.vertices {
			// dangling node
			if len(outlinks) == 0 {
				// distribute to all equally
				for page := range newRank {
					newRank[page] += damping * rank[url] / float64(N)
				}
			} else {
				share := rank[url] / float64(len(outlinks))
				for _, outlink := range outlinks {
					newRank[outlink] += damping * share
				}
			}
		}

		// copy newRank to rank
		for url := range graph.vertices {
			rank[url] = newRank[url]
		}

		log.Printf("Iteration %d done.", i+1)
	}

	return rank
}

func NormalizeScores(scores map[string]float64) map[string]float64 {
	max := 0.0
	min := 1000.0

	// get min and max scores
	for _, score := range scores {
		if score > max {
			max = score
		}
		if score < min {
			min = score
		}
	}

	if max == min {
		return scores // err, avoid quietly
	}

	norm := make(map[string]float64)

	// normalize
	for url, score := range scores {
		norm[url] = (score - min) / (max - min)
	}

	return norm
}

func SavePageRankScore(scores map[string]float64, collection *mongo.Collection, ctx context.Context) error {
	for url, score := range scores {
		filter := bson.M{"url": url}
		update := bson.M{
			"$set": bson.M{
				"url":   url,
				"score": score,
			},
		}

		opts := options.Update().SetUpsert(true)

		_, err := collection.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			return fmt.Errorf("Failed to upsert pagerank score for %s: %w", url, err)
		}
	}
	return nil
}
