package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	defaultHost    = "0.0.0.0"
	defaultPort    = 9617
	defaultTimeout = 5 * time.Second

	metricsEndpoint   = "/metrics"
	readinessEndpoint = "/readiness"
	livenessEndpoint  = "/liveness"

	piholeHostEnv  = "PIHOLE_HOST"
	piholeTokenEnv = "PIHOLE_API_TOKEN"
)

var (
	piholeHost     string
	piholeAPIToken string
)

func init() {
	host := os.Getenv(piholeHostEnv)
	if host == "" {
		log.Fatalln(piholeHostEnv, "must be set")
	}
	piholeHost = host

	token := os.Getenv(piholeTokenEnv)
	if token == "" {
		log.Fatalln(piholeTokenEnv, "must be set")
	}
	piholeAPIToken = token
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
		handler.ServeHTTP(w, r)

		updateMetrics(piholeHost, piholeAPIToken)
	})
}

func okHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
