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

func (s *PastebinScraper) handlePaste(paste pb.Paste, errChan chan error, wg *sync.WaitGroup) {
	s.sem <- struct{}{}
	defer func() {
		<-s.sem
		wg.Done()
	}()

	saved, err := s.storage.IsSaved(paste.Key)

	if err != nil {
		errChan <- err
		return
	}

	if !saved {
		pasteContent, err := s.api.GetPaste(paste.Key)
		if err != nil {
			errChan <- err
			return
		}
		paste.Content = string(pasteContent)
		err = s.storage.Save(paste)
		if err != nil {
			errChan <- err
			return
		}
		s.logger.Debugf("Saved %s", paste.Key)
	}
}

func (s *PastebinScraper) scrape() error {
	s.logger.Info("Started scraper")

	pastes, err := s.api.LatestPastes()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	errChan := make(chan error)
	done := make(chan struct{})

	for _, paste := range pastes {
		wg.Add(1)
		go s.handlePaste(paste, errChan, &wg)
	}

	go func() {
		wg.Wait()
		done <- struct{}{}
	}()

	select {
	case <-done:
		s.logger.Info("Ended scraper")
		return nil
	case err := <-errChan:
		return err
	}
}

// Start starts the scraping process and pings the Healthcheck endpoint.
func (s *PastebinScraper) Start() error {
	ticker := time.NewTicker(1 * time.Minute)
	for {
		s.logger.Info("Waiting for timer")
		<-ticker.C

		if err := s.healthcheck.Start(); err != nil {
			return err
		}

		if err := s.scrape(); err != nil {
			s.healthcheck.Fail(err.Error())
			return err
		}

		if err := s.healthcheck.Success(); err != nil {
			return err
		}
	}
}
