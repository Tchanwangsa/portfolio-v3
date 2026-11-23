package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"portfolio-v3/internal/runtimeenv"
	"portfolio-v3/internal/stats"
)

type statsAPI struct {
	dataFile     string
	tokenHash    [sha256.Size]byte
	origins      map[string]bool
	publishMutex sync.Mutex
	github       *githubStats
}

func main() {
	envFile := flag.String("env-file", "", "path to a KEY=value settings file")
	flag.Parse()
	env, err := runtimeenv.Load(*envFile)
	if err != nil {
		log.Fatal(err)
	}
	addr := env("STATS_LISTEN_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8787"
		if port := env("PORT"); port != "" {
			addr = net.JoinHostPort("", port)
		}
	}
	token := env("STATS_PUBLISH_TOKEN")
	if len(token) < 32 {
		log.Fatal("STATS_PUBLISH_TOKEN must contain at least 32 characters")
	}
	dataFile := env("STATS_DATA_FILE")
	if dataFile == "" {
		dataFile = "data/agentic-stats.json"
	}
	if err := os.MkdirAll(filepath.Dir(dataFile), 0o700); err != nil {
		log.Fatal(err)
	}
	api := &statsAPI{dataFile: dataFile, tokenHash: sha256.Sum256([]byte(token)), origins: make(map[string]bool)}
	for _, value := range strings.Split(env("STATS_ALLOWED_ORIGINS"), ",") {
		if origin := strings.TrimSpace(value); origin != "" {
			api.origins[origin] = true
		}
	}

	githubUser, githubToken := env("GITHUB_USERNAME"), env("GITHUB_TOKEN")
	if githubUser != "" || githubToken != "" {
		if githubUsername.MatchString(githubUser) && githubToken != "" {
			api.github = &githubStats{username: githubUser, token: githubToken}
		} else {
			log.Print("GitHub stats require a valid GITHUB_USERNAME and GITHUB_TOKEN")
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET "+stats.Path, api.get)
	mux.HandleFunc("GET "+githubStatsPath, api.getGitHub)
	mux.HandleFunc("PUT "+stats.Path, api.publish)
	server := &http.Server{
		Addr: addr, Handler: mux, ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second, IdleTimeout: 30 * time.Second,
	}
	log.Printf("portfolio stats API listening on %s", addr)
	log.Fatal(server.ListenAndServe())
}

func (api *statsAPI) setReadHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Vary", "Origin")
	if origin := r.Header.Get("Origin"); api.origins[origin] {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}
	w.Header().Set("Cache-Control", "no-store")
}

func (api *statsAPI) get(w http.ResponseWriter, r *http.Request) {
	api.setReadHeaders(w, r)
	data, err := os.ReadFile(api.dataFile)
	if errors.Is(err, os.ErrNotExist) {
		http.Error(w, "No stats have been published yet", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("read stats: %v", err)
		http.Error(w, "Stats are unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

func (api *statsAPI) publish(w http.ResponseWriter, r *http.Request) {
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	hash := sha256.Sum256([]byte(token))
	if !ok || subtle.ConstantTimeCompare(hash[:], api.tokenHash[:]) != 1 {
		w.Header().Set("WWW-Authenticate", "Bearer")
		http.Error(w, "Invalid publish token", http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	decoder.DisallowUnknownFields()
	var snapshot stats.Snapshot
	if err := decoder.Decode(&snapshot); err != nil || decoder.Decode(new(any)) != io.EOF {
		http.Error(w, "Expected one stats snapshot", http.StatusBadRequest)
		return
	}
	if err := snapshot.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	api.publishMutex.Lock()
	defer api.publishMutex.Unlock()
	if current, err := os.ReadFile(api.dataFile); err == nil {
		var previous stats.Snapshot
		if json.Unmarshal(current, &previous) == nil && previous.AsOf > snapshot.AsOf {
			http.Error(w, "A newer snapshot is already published", http.StatusConflict)
			return
		}
	}
	if err := snapshot.Write(api.dataFile); err != nil {
		log.Printf("persist stats: %v", err)
		http.Error(w, "Could not save stats", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
