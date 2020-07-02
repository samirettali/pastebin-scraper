package main

import (
	"os"

	healthcheck "github.com/samirettali/go-healthchecks"
	"github.com/samirettali/pastebin-scraper/scraper"
	"github.com/samirettali/pastebin-scraper/storage"
	log "github.com/sirupsen/logrus"
)

func main() {
	mongoURI := os.Getenv("MONGO_URI")
	mongoDatabase := os.Getenv("MONGO_DB")
	mongoCollection := os.Getenv("MONGO_COL")

	if mongoURI == "" || mongoDatabase == "" || mongoCollection == "" {
		log.Fatal("You must set MongoDB environment variables")
	}

	healthcheckURL := os.Getenv("HEALTHCHECK")

	if healthcheckURL == "" {
		log.Fatal("You must set HEALTHCHECK environment variable")
	}

	logger := log.New()
	logger.SetReportCaller(true)

	storage := &storage.MongoStorage{
		URI:        mongoURI,
		Database:   mongoDatabase,
		Collection: mongoCollection,
	}

	healthcheck := healthcheck.NewHealthcheck(healthcheckURL)

	scraper, err := scraper.NewScraper(8, storage, healthcheck, logger)
	if err != nil {
		logger.Fatal(err)
	}

	err = scraper.Start()
	if err != nil {
		logger.Fatal(err)
	}
}
