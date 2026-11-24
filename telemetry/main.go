// Command telemetry reports a rolling seven-day summary of local coding-agent
// token usage to the portfolio API.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"portfolio-v3/internal/runtimeenv"
	"portfolio-v3/internal/stats"
)

const usage = `Usage: telemetry <command> [flags]

Commands:
  run        Collect the last seven days of usage and publish it once
  install    Run the collector in the background on a schedule (macOS)
  uninstall  Stop the background collector
  status     Show the schedule, settings, and latest snapshot

Run "telemetry <command> -h" for a command's flags.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	commands := map[string]func([]string) error{
		"run": run, "install": install, "uninstall": uninstall, "status": status,
	}
	command, ok := commands[os.Args[1]]
	if !ok {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	if err := command(os.Args[2:]); err != nil {
		// The timestamp dates each failure in the background collector's log.
		fmt.Fprintf(os.Stderr, "%s telemetry %s: %v\n", time.Now().Format(time.DateTime), os.Args[1], err)
		os.Exit(1)
	}
}

// files locates everything the collector keeps outside the repository.
type files struct{ home string }

func locate() (files, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return files{}, err
	}
	return files{home: filepath.Join(dir, ".local", "share", "portfolio-telemetry")}, nil
}

func (f files) binary() string   { return filepath.Join(f.home, "telemetry") }
func (f files) settings() string { return filepath.Join(f.home, "settings.env") }
func (f files) snapshot() string { return filepath.Join(f.home, "agentic-stats.json") }
func (f files) log() string      { return filepath.Join(f.home, "telemetry.log") }

func run(args []string) error {
	paths, err := locate()
	if err != nil {
		return err
	}
	flags := flag.NewFlagSet("telemetry run", flag.ExitOnError)
	envFile := flags.String("env-file", "", "settings file (default: installed settings, if any)")
	tokscale := flags.String("tokscale", "", "path to tokscale binary (default: TOKSCALE_BIN, then PATH)")
	output := flags.String("output", paths.snapshot(), "path to save a local copy of the snapshot")
	flags.Parse(args)

	if *envFile == "" {
		if _, err := os.Stat(paths.settings()); err == nil {
			*envFile = paths.settings()
		}
	}
	snapshot, target, err := refresh(*envFile, *tokscale, *output)
	if err != nil {
		return err
	}
	fmt.Printf("%s: %d tokens in %d sessions, %s\n", snapshot.AsOf, snapshot.TotalTokens, snapshot.Sessions, target)
	return nil
}

// refresh collects a snapshot, saves it locally, and publishes it when an API
// URL is set, returning where the snapshot went.
func refresh(envFile, tokscale, output string) (stats.Snapshot, string, error) {
	var none stats.Snapshot
	env, err := runtimeenv.Load(envFile)
	if err != nil {
		return none, "", err
	}
	endpoint, token := env("STATS_PUBLISH_URL"), env("STATS_PUBLISH_TOKEN")
	if endpoint != "" {
		if err := checkTarget(endpoint, token); err != nil {
			return none, "", err
		}
	}
	if tokscale, err = findTokscale(tokscale, env); err != nil {
		return none, "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	snapshot, err := collect(ctx, tokscale, time.Now())
	if err != nil {
		return none, "", err
	}
	if err := save(snapshot, output); err != nil {
		return none, "", err
	}
	if endpoint == "" {
		return snapshot, "saved to " + output, nil
	}
	if err := publish(ctx, endpoint, token, snapshot); err != nil {
		return none, "", err
	}
	return snapshot, "published to " + endpoint, nil
}

func cmp(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}
