package main

import (
	"database/sql"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	piholeNamespace   = "pihole"
	exporterNamespace = "pihole_exporter"
)

var (
	DNSQueries          *prometheus.CounterVec
	AllowedDNSQueries   *prometheus.CounterVec
	BlockedDNSQueries   *prometheus.CounterVec
	ClientDNSQueries    *prometheus.CounterVec // queries with client label
	HTTPRequestDuration prometheus.Gauge
)

func buildMetrics() *prometheus.Registry {
	registry := prometheus.NewRegistry()

	DNSQueries = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: piholeNamespace,
			Name:      "dns_queries_total",
			Help:      "Total number of DNS queries with type labels",
		},
		[]string{"type"},
	)

	AllowedDNSQueries = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: piholeNamespace,
			Name:      "allowed_dns_queries",
			Help:      "Forwarded or cached DNS queries",
		},
		[]string{"status", "forwarded_to"},
	)

	BlockedDNSQueries = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: piholeNamespace,
			Name:      "blocked_dns_queries",
			Help:      "Blocked DNS queries",
		},
		[]string{"blocked_by", "deep_cname"},
	)

	ClientDNSQueries = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: piholeNamespace,
			Name:      "client_dns_queries",
			Help:      "Number of DNS queries with client labels",
		},
		[]string{"client"},
	)

	HTTPRequestDuration = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: exporterNamespace,
		Name:      "http_request_duration_seconds",
		Help:      "How long this exporter takes to respond when scraped by prometheus",
	})

	metrics := []prometheus.Collector{
		DNSQueries,
		AllowedDNSQueries,
		BlockedDNSQueries,
		ClientDNSQueries,
		HTTPRequestDuration,
	}

	for _, metric := range metrics {
		registry.MustRegister(metric)
	}

	return registry
}

func updateMetrics(piholeDB *sql.DB, since int64) int64 {
	now := time.Now().Unix()
	stats, err := queryPihole(piholeDB, since, now)
	if err != nil {
		log.Printf("Unable to query pihole database: %s", err)
		return since
	}

	for queryType, num := range stats.QueryTypes {
		DNSQueries.WithLabelValues(queryType).Add(num)
	}

	for status, upstreamMap := range stats.AllowedQueries {
		for upstream, num := range upstreamMap {
			AllowedDNSQueries.WithLabelValues(status, upstream).Add(num)
		}
	}

	for status, num := range stats.BlockedQueries {
		BlockedDNSQueries.WithLabelValues(status, "false").Add(num)
	}

	for status, num := range stats.BlockedCNAMEQueries {
		BlockedDNSQueries.WithLabelValues(status, "true").Add(num)
	}

	for client, num := range stats.ClientQueries {
		ClientDNSQueries.WithLabelValues(client).Add(num)
	}

	duration := float64(time.Now().Unix() - now)
	HTTPRequestDuration.Set(duration)

	return now
}
