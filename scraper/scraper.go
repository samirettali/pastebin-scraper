package scraper

import (
	"sync"
	"time"

	healthcheck "github.com/samirettali/go-healthchecks"
	pb "github.com/samirettali/go-pastebin"
	log "github.com/sirupsen/logrus"
)

// PastebinScraper is an interface that defines a scraper for Pastebin
type PastebinScraper struct {
	storage     Storage
	logger      *log.Logger
	healthcheck *healthcheck.Healthcheck
	sem         chan struct{}
	api         *pb.Client
}

func NewScraper(concurrency int, storage Storage, hc *healthcheck.Healthcheck, logger *log.Logger) (*PastebinScraper, error) {
	err := storage.Init()
	if err != nil {
		return nil, err
	}

	scraper := &PastebinScraper{
		storage:     storage,
		logger:      logger,
		healthcheck: hc,
		sem:         make(chan struct{}, concurrency),
		api:         pb.NewClient(),
	}

	return scraper, nil
}

// Storage is an interface that defines storage methods
type Storage interface {
	Init() error
	IsSaved(string) (bool, error)
	Save(pb.Paste) error
}

func (scraper *PastebinScraper) handlePaste(paste pb.Paste, wg *sync.WaitGroup) {
	scraper.sem <- struct{}{}
	defer func() {
		<-scraper.sem
		wg.Done()
	}()
	saved, err := scraper.storage.IsSaved(paste.Key)
	if err != nil {
		scraper.logger.Error(err)
		return
	}
	if !saved {
		pasteContent, err := scraper.api.GetPaste(paste.Key)
		if err != nil {
			scraper.logger.Error(err)
			return
		}
		paste.Content = string(pasteContent)
		err = scraper.storage.Save(paste)
		if err != nil {
			scraper.logger.Error(err)
			return
		}
		scraper.logger.Infof("Saved %s", paste.Key)
	}
}

func (scraper *PastebinScraper) scrape() error {
	scraper.logger.Info("Started scraper")
	wg := sync.WaitGroup{}
	pastes, err := scraper.api.LatestPastes()
	if err != nil {
		return err
	}
	for _, paste := range pastes {
		wg.Add(1)
		go scraper.handlePaste(paste, &wg)
	}
	wg.Wait()
	scraper.logger.Info("Ended scraper")
	return nil
}

// Start starts the scraping process
func (scraper *PastebinScraper) Start() error {
	// Goroutine that starts the scraping and pings healthcheck once a minute
	ticker := time.NewTicker(1 * time.Minute)
	for {
		scraper.logger.Info("Waiting for timer")
		<-ticker.C

		if err := scraper.healthcheck.Start(); err != nil {
			return err
		}

		if err := scraper.scrape(); err != nil {
			scraper.healthcheck.Fail(err.Error())
			return err
		}

		if err := scraper.healthcheck.Success(); err != nil {
			return err
		}
	}
}
