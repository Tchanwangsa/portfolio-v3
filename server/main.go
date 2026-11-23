package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"time"

	"portfolio-v3/internal/runtimeenv"
	"portfolio-v3/portfolioapi"
)

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
	handler, err := portfolioapi.New(env)
	if err != nil {
		log.Fatal(err)
	}
	server := &http.Server{
		Addr: addr, Handler: handler, ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second, IdleTimeout: 30 * time.Second,
	}
	log.Printf("portfolio stats API listening on %s", addr)
	log.Fatal(server.ListenAndServe())
}
