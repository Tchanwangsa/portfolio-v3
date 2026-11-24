package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"portfolio-v3/internal/stats"
)

// checkTarget rejects settings that would send the token somewhere unsafe.
func checkTarget(endpoint, token string) error {
	destination, err := url.Parse(endpoint)
	if err != nil || destination.Host == "" || destination.User != nil {
		return errors.New("invalid STATS_PUBLISH_URL")
	}
	host := destination.Hostname()
	local := host == "localhost" || net.ParseIP(host).IsLoopback()
	if destination.Scheme != "https" && !(destination.Scheme == "http" && local) {
		return errors.New("STATS_PUBLISH_URL requires HTTPS outside loopback")
	}
	if len(token) < 32 {
		return errors.New("STATS_PUBLISH_TOKEN must contain at least 32 characters")
	}
	return nil
}

func publish(ctx context.Context, endpoint, token string, snapshot stats.Snapshot) error {
	body, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	client := http.Client{
		Timeout: 10 * time.Second,
		// A redirect would forward the token to another host.
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	response, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("publish stats: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("publish stats: HTTP %d", response.StatusCode)
	}
	return nil
}
