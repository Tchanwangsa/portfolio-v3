package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"portfolio-v3/internal/runtimeenv"
	"portfolio-v3/internal/stats"
)

type graph struct {
	Contributions []struct {
		Date    string
		Totals  struct{ Tokens int64 }
		Clients []struct {
			ModelID string
			Tokens  struct{ Input, Output, CacheRead, CacheWrite, Reasoning int64 }
		}
	}
	TimeMetrics *struct{ SessionCount int }
}

func main() {
	home, _ := os.UserHomeDir()
	tokscale := flag.String("tokscale", filepath.Join(home, ".local", "share", "portfolio-v3-stats", "tokscale"), "path to the native Tokscale binary")
	output := flag.String("output", "public/agentic-stats.json", "path to the portfolio JSON feed")
	envFile := flag.String("env-file", "", "path to a KEY=value settings file")
	flag.Parse()

	if err := refresh(*tokscale, *output, *envFile); err != nil {
		fmt.Fprintln(os.Stderr, "Agentic coding stats refresh failed:", err)
		os.Exit(1)
	}
}

func refresh(tokscale, output, envFile string) error {
	env, err := runtimeenv.Load(envFile)
	if err != nil {
		return err
	}
	// Days follow this machine's local time zone, so they move with you when you travel.
	today := time.Now()
	start := today.AddDate(0, 0, -6)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, tokscale, "graph", "--offline", "--no-spinner",
		"--since", start.Format(time.DateOnly), "--until", today.Format(time.DateOnly))
	data, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("run Tokscale: %w", err)
	}
	var report graph
	if err := json.Unmarshal(data, &report); err != nil {
		return fmt.Errorf("decode Tokscale graph: %w", err)
	}
	if report.Contributions == nil {
		return fmt.Errorf("Tokscale graph is missing daily contributions")
	}

	days := make(map[string]int64, len(report.Contributions))
	models := make(map[string]int64)
	for _, day := range report.Contributions {
		days[day.Date] = day.Totals.Tokens
		for _, client := range day.Clients {
			tokens := client.Tokens
			models[client.ModelID] += tokens.Input + tokens.Output + tokens.CacheRead + tokens.CacheWrite + tokens.Reasoning
		}
	}

	result := stats.Snapshot{AsOf: today.Format(time.DateOnly)}
	for offset := range 7 {
		date := start.AddDate(0, 0, offset)
		tokens := days[date.Format(time.DateOnly)]
		result.DailyTokens = append(result.DailyTokens, stats.DayUsage{Day: date.Format("Mon"), Tokens: tokens})
		result.TotalTokens += tokens
	}
	if report.TimeMetrics != nil {
		result.Sessions = report.TimeMetrics.SessionCount
	} else if result.TotalTokens > 0 {
		return fmt.Errorf("Tokscale graph is missing session metrics")
	}

	var topModel string
	var topTokens int64
	for model, tokens := range models {
		if model != "" && tokens > 0 && (tokens > topTokens || (tokens == topTokens && model < topModel)) {
			topModel, topTokens = model, tokens
		}
	}
	if topModel != "" {
		result.MostUsedModel = &topModel
		result.MostUsedModelTokens = &topTokens
	}
	if err := result.Validate(); err != nil {
		return err
	}
	if err := result.Write(output); err != nil {
		return err
	}
	if endpoint := env("STATS_PUBLISH_URL"); endpoint != "" {
		return publish(endpoint, env("STATS_PUBLISH_TOKEN"), result)
	}
	return nil
}

func publish(endpoint, token string, snapshot stats.Snapshot) error {
	destination, err := url.Parse(endpoint)
	if err != nil || destination.Host == "" || destination.User != nil {
		return fmt.Errorf("invalid STATS_PUBLISH_URL")
	}
	host := destination.Hostname()
	local := host == "localhost" || net.ParseIP(host).IsLoopback()
	if destination.Scheme != "https" && !(destination.Scheme == "http" && local) {
		return fmt.Errorf("STATS_PUBLISH_URL requires HTTPS outside loopback")
	}
	if len(token) < 32 {
		return fmt.Errorf("STATS_PUBLISH_TOKEN must contain at least 32 characters")
	}
	body, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, destination.String(), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	client := http.Client{
		Timeout:       10 * time.Second,
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
