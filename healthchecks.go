package main

import (
	"bytes"
	"crypto/tls"
	"net/http"
	"time"
)

// Healthcheck is an implementation of a "client" for healthchecks.io
type Healthcheck struct {
	URL    string
	client *http.Client
}

func NewHealthcheck(URL string) *Healthcheck {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	client.Head("google.com")

	return &Healthcheck{
		URL,
		client,
	}
}

// Success just pings the Healtchecks endpoint
func (h *Healthcheck) Success() error {
	_, err := h.client.Head(h.URL)
	if err != nil {
		return err
	}

	return nil
}

// Fail sends a message to the Healtchecks endpoint
func (h *Healthcheck) Fail(msg string) error {
	url := h.URL + "/fail"
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(msg))
	if err != nil {
		return err
	}

	_, err = h.client.Do(req)
	if err != nil {
		return err
	}

	return nil
}

// Start send a ping to the start endpoint
func (h *Healthcheck) Start() error {
	url := h.URL + "/start"
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return err
	}

	_, err = h.client.Do(req)
	if err != nil {
		return err
	}

	return nil
}
