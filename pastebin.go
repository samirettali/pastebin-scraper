package main

import (
	"encoding/json"
	"fmt"
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
	msgChan     chan string
	errChan     chan error
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
		scraper.errChan <- err
		return
	}
	if !saved {
		url := "https://scrape.pastebin.com/api_scrape_item.php?i=" + paste.Key
		pasteContent, err := makeRequest(url)
		if err != nil {
			scraper.errChan <- err
			return
		}
		paste.Content = string(pasteContent)
		err = scraper.storage.Save(paste)
		if err != nil {
			scraper.errChan <- err
			return
		}
		scraper.msgChan <- fmt.Sprintf("Saved %s", paste.Key)
	}
}

func (scraper *PastebinScraper) scrape() {
	wg := sync.WaitGroup{}
	sem := make(chan bool, scraper.concurrency)
	pastes, err := scraper.getPastes()
	if err != nil {
		scraper.errChan <- err
		return
	}
	for _, paste := range pastes {
		wg.Add(1)
		go scraper.handlePaste(paste, sem, &wg)
	}
	wg.Wait()
}

// Start starts the scraping process
func (scraper *PastebinScraper) Start() error {
	scraper.msgChan = make(chan string, scraper.concurrency)
	scraper.errChan = make(chan error, scraper.concurrency)

	err := scraper.storage.Init()
	if err != nil {
		return err
	}

	healthcheck := NewHealthcheck(os.Getenv("HEALTHCHECK"))

	// Goroutine that starts the scraping and pings healthcheck once a minute
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		for {
			scraper.msgChan <- "Waiting for timer"
			<-ticker.C
			err := healthcheck.Start()
			if err != nil {
				scraper.errChan <- err
				return
			}
			scraper.msgChan <- "Started scraper"
			scraper.scrape()
			scraper.msgChan <- "Scraper ended"
			err = healthcheck.Success()
			if err != nil {
				scraper.errChan <- err
				return
			}
		}
	}()

	for {
		select {
		case msg := <-scraper.msgChan:
			log.Info(msg)
		case err := <-scraper.errChan:
			log.Error(err)
			healthcheck.Fail(err.Error())
			return err
		}
	}
}
