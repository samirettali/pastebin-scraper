package main

import (
	"bytes"
	"net/http"
	"time"
)

// Healthcheck is an implementation of a "client" for healthchecks.io
type Healthcheck struct {
	URL string
}

// Success just pings the Healtchecks endpoint
func (h *Healthcheck) Success() error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	_, err := client.Head(h.URL)
	if err != nil {
		return err
	}

	return nil
}

// Fail sends a message to the Healtchecks endpoint
func (h *Healthcheck) Fail(msg string) error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	url := h.URL + "/fail"
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(msg))
	if err != nil {
		return err
	}

	_, err = client.Do(req)
	if err != nil {
		return err
	}

	return nil
}

// Start send a ping to the start endpoint
func (h *Healthcheck) Start() error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	url := h.URL + "/start"
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return err
	}

	_, err = client.Do(req)
	if err != nil {
		return err
	}

	return nil
}
