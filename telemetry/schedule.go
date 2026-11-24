package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"portfolio-v3/internal/runtimeenv"
	"portfolio-v3/internal/stats"
)

// label names the launchd agent that runs the collector.
const label = "xyz.tchanwangsa.portfolio.telemetry"

var errNeedsMacOS = errors.New("scheduling requires macOS (launchd); elsewhere, run `telemetry run` from cron")

func install(args []string) error {
	paths, err := locate()
	if err != nil {
		return err
	}
	flags := flag.NewFlagSet("telemetry install", flag.ExitOnError)
	envFile := flags.String("env-file", "", "settings file to install (default: keep installed settings)")
	tokscale := flags.String("tokscale", "", "path to tokscale binary (default: TOKSCALE_BIN, then PATH)")
	interval := flags.Duration("interval", 15*time.Minute, "time between runs")
	flags.Parse(args)
	if runtime.GOOS != "darwin" {
		return errNeedsMacOS
	}
	if *interval < time.Minute {
		return errors.New("-interval must be at least 1m")
	}

	if *envFile != "" {
		data, err := os.ReadFile(*envFile)
		if err != nil {
			return err
		}
		if err := writeFile(paths.settings(), data, 0o600); err != nil {
			return err
		}
	} else if !exists(paths.settings()) {
		return errors.New("no settings installed yet: copy telemetry/.env.example, fill it in, and pass -env-file")
	}
	env, err := runtimeenv.Load(paths.settings())
	if err != nil {
		return err
	}
	if env("STATS_PUBLISH_URL") == "" {
		return fmt.Errorf("STATS_PUBLISH_URL is not set in %s", paths.settings())
	}

	// A first run in the foreground catches bad settings before anything is scheduled.
	snapshot, target, err := refresh(paths.settings(), *tokscale, paths.snapshot())
	if err != nil {
		return err
	}
	// The agent gets an absolute path because launchd does not search the shell's PATH.
	tokscalePath, err := findTokscale(*tokscale, env)
	if err != nil {
		return err
	}
	if err := copySelf(paths.binary()); err != nil {
		return err
	}
	if err := stop(); err != nil {
		return err
	}
	if err := writeFile(agentFile(), agent(paths, tokscalePath, *interval), 0o644); err != nil {
		return err
	}
	if err := launchctl("bootstrap", fmt.Sprintf("gui/%d", os.Getuid()), agentFile()); err != nil {
		return err
	}
	fmt.Printf("%s: %d tokens in %d sessions, %s\n", snapshot.AsOf, snapshot.TotalTokens, snapshot.Sessions, target)
	fmt.Printf("Installed: the collector now runs every %s and at login.\n", strings.TrimSuffix(interval.String(), "0s"))
	return nil
}

func uninstall(args []string) error {
	paths, err := locate()
	if err != nil {
		return err
	}
	flags := flag.NewFlagSet("telemetry uninstall", flag.ExitOnError)
	purge := flags.Bool("purge", false, "also delete settings, log, and local snapshot")
	flags.Parse(args)
	if runtime.GOOS != "darwin" {
		return errNeedsMacOS
	}

	if err := stop(); err != nil {
		return err
	}
	for _, path := range []string{agentFile(), paths.binary()} {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	if *purge {
		if err := os.RemoveAll(paths.home); err != nil {
			return err
		}
		fmt.Println("Uninstalled and deleted", paths.home)
		return nil
	}
	fmt.Println("Uninstalled. Settings stay in", paths.home, "so `telemetry install` resumes without flags.")
	return nil
}

func status([]string) error {
	paths, err := locate()
	if err != nil {
		return err
	}
	schedule := "not installed"
	if scheduled() {
		schedule = "installed (" + label + ")"
	}
	fmt.Println("Schedule:  ", schedule)

	settings := "none"
	if exists(paths.settings()) {
		settings = paths.settings()
		if env, err := runtimeenv.Load(settings); err == nil && env("STATS_PUBLISH_URL") != "" {
			settings += ", publishing to " + env("STATS_PUBLISH_URL")
		}
	}
	fmt.Println("Settings:  ", settings)

	var saved time.Time
	latest := "none"
	if info, err := os.Stat(paths.snapshot()); err == nil {
		saved = info.ModTime()
		var snapshot stats.Snapshot
		if data, err := os.ReadFile(paths.snapshot()); err == nil && json.Unmarshal(data, &snapshot) == nil {
			latest = fmt.Sprintf("%s, %d tokens in %d sessions, saved %s ago",
				snapshot.AsOf, snapshot.TotalTokens, snapshot.Sessions, time.Since(saved).Round(time.Second))
		}
	}
	fmt.Println("Snapshot:  ", latest)

	// The log only grows on failures, so a log newer than the snapshot means the last run failed.
	if info, err := os.Stat(paths.log()); err == nil && info.ModTime().After(saved) {
		data, _ := os.ReadFile(paths.log())
		if lines := strings.Split(strings.TrimSpace(string(data)), "\n"); lines[len(lines)-1] != "" {
			fmt.Println("Last error:", lines[len(lines)-1])
		}
	}
	return nil
}

func agentFile() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", label+".plist")
}

// agent renders the launchd property list for the scheduled collector.
func agent(paths files, tokscale string, interval time.Duration) []byte {
	text := func(value string) string {
		var escaped bytes.Buffer
		_ = xml.EscapeText(&escaped, []byte(value))
		return "<string>" + escaped.String() + "</string>"
	}
	var arguments strings.Builder
	for _, argument := range []string{paths.binary(), "run", "-env-file", paths.settings(), "-tokscale", tokscale} {
		arguments.WriteString("\n\t\t" + text(argument))
	}
	return []byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	` + text(label) + `
	<key>ProgramArguments</key>
	<array>` + arguments.String() + `
	</array>
	<key>EnvironmentVariables</key>
	<dict>
		<key>PATH</key>
		` + text(os.Getenv("PATH")) + `
	</dict>
	<key>RunAtLoad</key>
	<true/>
	<key>StartInterval</key>
	<integer>` + fmt.Sprint(int(interval.Seconds())) + `</integer>
	<key>ProcessType</key>
	<string>Background</string>
	<key>StandardOutPath</key>
	<string>/dev/null</string>
	<key>StandardErrorPath</key>
	` + text(paths.log()) + `
</dict>
</plist>
`)
}

// copySelf persists the running binary so launchd does not point at a temporary `go run` build.
func copySelf(destination string) error {
	self, err := os.Executable()
	if err != nil {
		return err
	}
	if self, err = filepath.EvalSymlinks(self); err != nil {
		return err
	}
	if installed, err := filepath.EvalSymlinks(destination); err == nil && installed == self {
		return nil
	}
	data, err := os.ReadFile(self)
	if err != nil {
		return err
	}
	return writeFile(destination, data, 0o755)
}

func service() string { return fmt.Sprintf("gui/%d/%s", os.Getuid(), label) }

func scheduled() bool {
	return runtime.GOOS == "darwin" && exec.Command("launchctl", "print", service()).Run() == nil
}

// stop unloads the agent and waits, because bootout returns before launchd finishes.
func stop() error {
	if !scheduled() {
		return nil
	}
	if err := launchctl("bootout", service()); err != nil {
		return err
	}
	for range 50 {
		if !scheduled() {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return errors.New("launchd did not unload the collector")
}

func launchctl(args ...string) error {
	output, err := exec.Command("launchctl", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("launchctl %s: %w: %s", args[0], err, strings.TrimSpace(string(output)))
	}
	return nil
}
