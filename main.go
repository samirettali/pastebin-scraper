package main

import (
	"os"

	"github.com/kelseyhightower/envconfig"
	healthcheck "github.com/samirettali/go-healthchecks"
	"github.com/samirettali/pastebin-scraper/scraper"
	"github.com/samirettali/pastebin-scraper/storage"
	log "github.com/sirupsen/logrus"
)

func main() {
	storageType, found := os.LookupEnv("STORAGE_TYPE")

	if !found {
		log.Fatal("You have to set STORAGE_TYPE env variable")
	}

	var store scraper.Storage

	switch storageType {
	case "mongo":
		var c storage.MongoConfig
		err := envconfig.Process("mongo", &c)
		if err != nil {
			log.Fatal(err.Error())
		}
		store = &storage.MongoStorage{
			Config: &c,
		}
	case "postgres":
		var c storage.PgConfig
		err := envconfig.Process("postgres", &c)
		if err != nil {
			log.Fatal(err.Error())
		}
		store = &storage.PgStorage{
			Config: &c,
		}

	default:
		log.Fatal("You have to set STORAGE_TYPE env variable")
	}

	healthcheckURL := os.Getenv("HEALTHCHECK")

	if healthcheckURL == "" {
		log.Fatal("You must set a valid HEALTHCHECK environment variable")
	}

	logger := log.New()
	logger.SetReportCaller(true)

	healthcheck := healthcheck.NewHealthcheck(healthcheckURL)

	scraper, err := scraper.NewScraper(8, store, healthcheck, logger)
	if err != nil {
		logger.Fatal(err)
	}

	err = scraper.Start()
	if err != nil {
		logger.Fatal(err)
	}
}
