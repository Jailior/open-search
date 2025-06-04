package main

import (
	"log"

	"github.com/Jailior/open-search/backend/internal/pagerank"
	"github.com/Jailior/open-search/backend/internal/storage"
)

const DAMPING_FACTOR = 0.85
const ITERATIONS = 20

const DB_NAME = "opensearch"
const RAW_PAGE_COLL = "pages"
const PAGE_RANK_COLL = "pagerank"

func main() {

	db := storage.MakeDB()
	db.Connect()
	defer db.Disconnect()

	db.AddCollection(DB_NAME, RAW_PAGE_COLL)
	collection := db.GetCollection(RAW_PAGE_COLL)

	graph := pagerank.MakeGraph()
	graph.BuildFromPages(collection, *db.GetContext())

	ranks := pagerank.PageRank(graph, DAMPING_FACTOR, ITERATIONS)
	normRanks := pagerank.NormalizeScores(ranks)

	db.AddCollection(DB_NAME, PAGE_RANK_COLL)
	pageRankCollection := db.GetCollection(PAGE_RANK_COLL)

	err := pagerank.SavePageRankScore(normRanks, pageRankCollection, *db.GetContext())
	if err != nil {
		log.Fatal("Error saving PageRank scores: ", err)
	}
}
