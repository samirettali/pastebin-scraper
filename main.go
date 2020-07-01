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

	logger := log.New()
	logger.SetReportCaller(true)

	scraper, err := NewScraper(8, storage, logger)

	if err != nil {
		logger.Fatal(err)
	}

	err = scraper.Start()
	if err != nil {
		log.Fatal(err)
	}
}
