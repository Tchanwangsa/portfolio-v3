package handler

import (
	"log"
	"net/http"
	"os"
	"sync"

	"portfolio-v3/portfolioapi"
)

var setup = sync.OnceValues(func() (http.Handler, error) {
	return portfolioapi.New(os.Getenv)
})

func Handler(w http.ResponseWriter, r *http.Request) {
	mux, err := setup()
	if err != nil {
		log.Printf("configure portfolio API: %v", err)
		http.Error(w, "Portfolio API is unavailable", http.StatusServiceUnavailable)
		return
	}
	if r.URL.Path == "/api/portfolio" {
		path := r.URL.Query().Get("__portfolio_route")
		switch path {
		case "/healthz", "/api/agentic-stats", "/api/github-stats":
			r.URL.Path = path
			r.URL.RawPath = ""
		default:
			http.NotFound(w, r)
			return
		}
	}
	mux.ServeHTTP(w, r)
}
