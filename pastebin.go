package main

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// Paste is a struct that represents a paste object from Pastebin's API.
// I excluded full_url and scrape_url to reduce space usage, as they can be
// derived from the paste key
type Paste struct {
	Date    string `json:"date"`
	Key     string `json:"key"`
	Expire  string `json:"expire"`
	Title   string `json:"title"`
	Syntax  string `json:"syntax"`
	User    string `json:"user"`
	Content string `json:"content"`
}

// PastebinScraper is an interface that defines a scraper for Pastebin
type PastebinScraper struct {
	concurrency int
	storage     Storage
	logger      *log.Logger
	healthcheck *Healthcheck
}

func NewScraper(concurrency int, storage Storage, logger *log.Logger) (*PastebinScraper, error) {
	err := storage.Init()
	if err != nil {
		return nil, err
	}

	healthcheck := NewHealthcheck(os.Getenv("HEALTHCHECK"))

	return &PastebinScraper{
		concurrency,
		storage,
		logger,
		healthcheck,
	}, nil
}

// Storage is an interface that defines storage methods
type Storage interface {
	Init() error
	IsSaved(string) (bool, error)
	Save(Paste) error
}

func (scraper *PastebinScraper) getPastes() ([]Paste, error) {
	const scrapeURL = "https://scrape.pastebin.com/api_scraping.php?limit=250"
	rawPastes, err := makeRequest(scrapeURL)
	if err != nil {
		return nil, err
	}
	pastes := make([]Paste, 0)
	err = json.Unmarshal(rawPastes, &pastes)
	if err != nil {
		return nil, err
	}
	return pastes, nil
}

func (scraper *PastebinScraper) handlePaste(paste Paste, sem chan bool, wg *sync.WaitGroup) {
	sem <- true
	defer func() {
		<-sem
		wg.Done()
	}()
	saved, err := scraper.storage.IsSaved(paste.Key)
	if err != nil {
		scraper.logger.Error(err)
		scraper.healthcheck.Fail(err.Error())
		return
	}
	if !saved {
		url := "https://scrape.pastebin.com/api_scrape_item.php?i=" + paste.Key
		pasteContent, err := makeRequest(url)
		if err != nil {
			scraper.logger.Error(err)
			scraper.healthcheck.Fail(err.Error())
			return
		}
		paste.Content = string(pasteContent)
		err = scraper.storage.Save(paste)
		if err != nil {
			scraper.logger.Error(err)
			scraper.healthcheck.Fail(err.Error())
			return
		}
		scraper.logger.Infof("Saved %s", paste.Key)
	}
}

func (scraper *PastebinScraper) scrape() error {
	wg := sync.WaitGroup{}
	sem := make(chan bool, scraper.concurrency)
	pastes, err := scraper.getPastes()
	if err != nil {
		return err
	}
	for _, paste := range pastes {
		wg.Add(1)
		go scraper.handlePaste(paste, sem, &wg)
	}
	wg.Wait()
	return nil
}

// Start starts the scraping process
func (scraper *PastebinScraper) Start() error {

	// Goroutine that starts the scraping and pings healthcheck once a minute
	ticker := time.NewTicker(1 * time.Minute)
	for {
		scraper.logger.Info("Waiting for timer")
		<-ticker.C
		err := scraper.healthcheck.Start()
		if err != nil {
			scraper.logger.Error(err)
			return err
		}
		scraper.logger.Info("Started scraper")
		err = scraper.scrape()
		if err != nil {
			scraper.healthcheck.Fail(err.Error())
			scraper.logger.Error(err)
			return err
		}
		scraper.logger.Info("Scraper ended")
		err = scraper.healthcheck.Success()
		if err != nil {
			scraper.logger.Error(err)
			return err
		}
	}
}
