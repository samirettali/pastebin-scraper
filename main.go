package main

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func main() {
	mongoURI := os.Getenv("MONGO_URI")
	mongoDatabase := os.Getenv("MONGO_DB")
	mongoCollection := os.Getenv("MONGO_COL")

	if mongoURI == "" || mongoDatabase == "" || mongoCollection == "" {
		log.Fatal("You must set environment variables")
	}

	storage := &MongoStorage{
		URI:        mongoURI,
		Database:   mongoDatabase,
		Collection: mongoCollection,
	}

	scraper := &PastebinScraper{
		concurrency: 8,
		storage:     storage,
	}

	err := scraper.Start()
	if err != nil {
		log.Error(err)
	}
}
