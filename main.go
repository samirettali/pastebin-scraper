package main

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {
	mongoURI := os.Getenv("MONGO_URI")
	mongoDatabase := os.Getenv("MONGO_DB")
	mongoCollection := os.Getenv("MONGO_COL")

	if mongoURI == "" || mongoDatabase == "" || mongoCollection == "" {
		log.Fatal("You must set environment variables")
	}

	log.SetLevel(log.DebugLevel)

	log.Debugf("MONGO_URI: %s", mongoURI)
	log.Debugf("MONGO_DB: %s", mongoDatabase)
	log.Debugf("MONGO_COL: %s", mongoCollection)

	storage := &MongoStorage{
		URI:        mongoURI,
		Database:   mongoDatabase,
		Collection: mongoCollection,
	}

	crawler := &PastebinScraper{
		concurrency: 8,
		storage:     storage,
	}

	err := crawler.Start()
	if err != nil {
		log.Error(err)
	}
}
