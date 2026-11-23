package portfolioapi

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"portfolio-v3/internal/stats"
)

type statsAPI struct {
	data      document
	tokenHash [sha256.Size]byte
	origins   map[string]bool
	github    *githubStats
}

func New(env func(string) string) (http.Handler, error) {
	token := env("STATS_PUBLISH_TOKEN")
	if len(token) < 32 {
		return nil, errors.New("STATS_PUBLISH_TOKEN must contain at least 32 characters")
	}
	statsFile := env("STATS_DATA_FILE")
	if statsFile == "" {
		statsFile = "data/agentic-stats.json"
	}
	statsData, err := openDocument(env, statsFile, "agentic-stats.json")
	if err != nil {
		return nil, err
	}
	api := &statsAPI{data: statsData, tokenHash: sha256.Sum256([]byte(token)), origins: make(map[string]bool)}
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
	mux.HandleFunc("PUT "+stats.Path, api.publish)
	mux.HandleFunc("GET "+githubStatsPath, api.getGitHub)
	if webhookToken := env("RUNS_WEBHOOK_TOKEN"); webhookToken != "" {
		if len(webhookToken) < 32 || webhookToken == token {
			return nil, errors.New("set a distinct RUNS_WEBHOOK_TOKEN of at least 32 characters")
		}
		runFile := filepath.Join(filepath.Dir(statsFile), "running-stats.json")
		if runFile == filepath.Clean(statsFile) {
			return nil, errors.New("STATS_DATA_FILE conflicts with the running stats file")
		}
		runData, err := openDocument(env, runFile, "running-stats.json")
		if err != nil {
			return nil, err
		}
		runs := &runningAPI{data: runData, tokenHash: sha256.Sum256([]byte(webhookToken))}
		mux.HandleFunc("GET "+runsPath, func(w http.ResponseWriter, r *http.Request) {
			api.setReadHeaders(w, r)
			runs.get(w, r)
		})
		mux.HandleFunc("POST "+runsPath, runs.receive)
	}
	return mux, nil
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
	data, err := api.data.read(r.Context())
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

func hasBearerToken(r *http.Request, expected [sha256.Size]byte) bool {
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	hash := sha256.Sum256([]byte(token))
	return ok && subtle.ConstantTimeCompare(hash[:], expected[:]) == 1
}

func (api *statsAPI) publish(w http.ResponseWriter, r *http.Request) {
	if !hasBearerToken(r, api.tokenHash) {
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

	stale := errors.New("a newer snapshot is already published")
	err := api.data.update(r.Context(), func(current []byte) ([]byte, error) {
		var previous stats.Snapshot
		if current != nil && json.Unmarshal(current, &previous) == nil && previous.AsOf > snapshot.AsOf {
			return nil, stale
		}
		return json.Marshal(snapshot)
	})
	if errors.Is(err, stale) {
		http.Error(w, "A newer snapshot is already published", http.StatusConflict)
		return
	}
	if err != nil {
		log.Printf("persist stats: %v", err)
		http.Error(w, "Could not save stats", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
