package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"portfolio-v3/internal/stats"
)

// graph is the part of `tokscale graph` output that the snapshot needs.
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

// findTokscale resolves tokscale from path, TOKSCALE_BIN, or PATH.
func findTokscale(path string, env func(string) string) (string, error) {
	path = cmp(cmp(path, env("TOKSCALE_BIN")), "tokscale")
	found, err := exec.LookPath(path)
	if err != nil {
		return "", fmt.Errorf("tokscale not found (%w): install it with `npm install -g tokscale` or set TOKSCALE_BIN", err)
	}
	return filepath.Abs(found)
}

// collect asks tokscale for the seven days ending today. Days follow the
// machine's local time zone, which tokscale also uses.
func collect(ctx context.Context, tokscale string, today time.Time) (stats.Snapshot, error) {
	start := today.AddDate(0, 0, -6)
	cmd := exec.CommandContext(ctx, tokscale, "graph", "--offline", "--no-spinner",
		"--since", start.Format(time.DateOnly), "--until", today.Format(time.DateOnly))
	data, err := cmd.Output()
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) && len(exit.Stderr) > 0 {
			return stats.Snapshot{}, fmt.Errorf("run tokscale: %w: %s", err, strings.TrimSpace(string(exit.Stderr)))
		}
		return stats.Snapshot{}, fmt.Errorf("run tokscale: %w", err)
	}
	var report graph
	if err := json.Unmarshal(data, &report); err != nil {
		return stats.Snapshot{}, fmt.Errorf("decode tokscale graph: %w", err)
	}
	return summarize(report, today)
}

func summarize(report graph, today time.Time) (stats.Snapshot, error) {
	if report.Contributions == nil {
		return stats.Snapshot{}, errors.New("tokscale graph is missing daily contributions")
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

	snapshot := stats.Snapshot{AsOf: today.Format(time.DateOnly)}
	for offset := -6; offset <= 0; offset++ {
		date := today.AddDate(0, 0, offset)
		tokens := days[date.Format(time.DateOnly)]
		snapshot.DailyTokens = append(snapshot.DailyTokens, stats.DayUsage{Day: date.Format("Mon"), Tokens: tokens})
		snapshot.TotalTokens += tokens
	}
	if report.TimeMetrics != nil {
		snapshot.Sessions = report.TimeMetrics.SessionCount
	} else if snapshot.TotalTokens > 0 {
		return stats.Snapshot{}, errors.New("tokscale graph is missing session metrics")
	}

	// Ties go to the alphabetically first model so repeated runs agree.
	var topModel string
	var topTokens int64
	for model, tokens := range models {
		if model != "" && tokens > 0 && (tokens > topTokens || (tokens == topTokens && model < topModel)) {
			topModel, topTokens = model, tokens
		}
	}
	if topModel != "" {
		snapshot.MostUsedModel = &topModel
		snapshot.MostUsedModelTokens = &topTokens
	}
	return snapshot, snapshot.Validate()
}

// save replaces the local copy atomically.
func save(snapshot stats.Snapshot, path string) error {
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(path, append(data, '\n'), 0o600)
}

func writeFile(path string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".telemetry-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	_, writeErr := file.Write(data)
	if err := errors.Join(writeErr, file.Chmod(mode), file.Close()); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}
