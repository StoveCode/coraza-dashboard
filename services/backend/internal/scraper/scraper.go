package scraper

import (
	"context"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// Scraper periodically fetches Prometheus metrics from coraza-spoa.
type Scraper struct {
	url    string
	mu     sync.RWMutex
	latest string
}

func New(url string) *Scraper {
	return &Scraper{url: url}
}

// Run scrapes the metrics endpoint every 15 seconds.
func (s *Scraper) Run(ctx context.Context) {
	log.Info().Str("url", s.url).Msg("starting Prometheus scraper")
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	// Scrape once immediately
	s.scrape()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.scrape()
		}
	}
}

func (s *Scraper) scrape() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, s.url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Warn().Err(err).Str("url", s.url).Msg("metrics scrape failed")
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		log.Warn().Err(err).Msg("failed to read metrics body")
		return
	}

	s.mu.Lock()
	s.latest = string(body)
	s.mu.Unlock()
	log.Debug().Int("bytes", len(body)).Msg("scraped metrics")
}

// Latest returns the most recently scraped metrics text (Prometheus format).
func (s *Scraper) Latest() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.latest
}
