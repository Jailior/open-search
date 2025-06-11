package main

import (
	"log"

	"github.com/Jailior/open-search/backend/internal/pagerank"
	"github.com/Jailior/open-search/backend/internal/storage"
)

// Damping factor preventing page rank scores exploding to inf, standard value is 0.85
const DAMPING_FACTOR = 0.85

// Number of iterations across pages for pagerank distribution
const ITERATIONS = 20

// Mongo database name
const DB_NAME = "opensearch"

// Raw pages collection
const RAW_PAGE_COLL = "pages"

// PageRank scores collection
const PAGE_RANK_COLL = "pagerank"

/*
Scores all pages in raw page collection based on url authority
and stores results in pagerank scores collection
*/
func main() {

	// connect to database
	db := storage.MakeDB()
	db.Connect()
	defer db.Disconnect()

	// add raw page collection and get its reference
	db.AddCollection(DB_NAME, RAW_PAGE_COLL)
	collection := db.GetCollection(RAW_PAGE_COLL)

	// make and build graph representing corpus from raw pages
	graph := pagerank.MakeGraph()
	graph.BuildFromPages(collection, *db.GetContext())

	// run pagerank algorithm on graph of raw pages
	ranks := pagerank.PageRank(graph, DAMPING_FACTOR, ITERATIONS)

	// normalize pagerank scores
	normRanks := pagerank.NormalizeScores(ranks)

	// add pagerank collection and get its reference
	db.AddCollection(DB_NAME, PAGE_RANK_COLL)
	pageRankCollection := db.GetCollection(PAGE_RANK_COLL)

	// make Mongo index on "url" field
	db.MakeIndex(PAGE_RANK_COLL, "url")

	// store PageRank scores in pagerank collection
	err := pagerank.SavePageRankScore(normRanks, pageRankCollection, *db.GetContext())
	if err != nil {
		log.Fatal("Error saving PageRank scores: ", err)
	}
}
