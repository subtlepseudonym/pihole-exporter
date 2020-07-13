package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	defaultHost    = "0.0.0.0"
	defaultPort    = 9617
	defaultTimeout = 5 * time.Second

	metricsEndpoint   = "/metrics"
	readinessEndpoint = "/readiness"
	livenessEndpoint  = "/liveness"

	sqlite3Driver = "sqlite3"
	piholeDSNEnv  = "PIHOLE_DSN"
)

var (
	lastUpdate int64
	piholeDB   *sql.DB
)

func init() {
	piholeDSN := os.Getenv(piholeDSNEnv)
	if piholeDSN == "" {
		log.Fatalln(piholeDSNEnv, "must be set")
	}

	var err error
	piholeDB, err = sql.Open(sqlite3Driver, piholeDSN)
	if err != nil {
		log.Fatalf("open db connection: %s", err)
	}
}

func main() {
	registry := buildMetrics()
	handlerOpts := promhttp.HandlerOpts{
		Registry: registry,
		Timeout:  defaultTimeout,
	}
	promHandler := promhttp.HandlerFor(registry, handlerOpts)

	mux := http.NewServeMux()
	mux.Handle(metricsEndpoint, metricsHandler(promHandler))
	mux.HandleFunc(readinessEndpoint, okHandler)
	mux.HandleFunc(livenessEndpoint, okHandler)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", defaultHost, defaultPort),
		Handler: mux,
	}

	log.Printf("Listening at %s:%d", defaultHost, defaultPort)
	log.Fatal(srv.ListenAndServe())
}

func metricsHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		unchanged := lastUpdate
		lastUpdate = updateMetrics(piholeDB, lastUpdate)
		if lastUpdate == unchanged {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

func okHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
