package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "pihole"

var (
	blockedDomains      prometheus.Gauge
	dnsQueries          *DailyCounter
	blockedAds          *DailyCounter
	forwardedQueries    *DailyCounter
	cachedQueries       *DailyCounter
	uniqueDomains       *DailyCounter
	clients             *DailyCounter
	uniqueClients       *DailyCounter
	replies             *prometheus.GaugeVec
	topDomains          *prometheus.GaugeVec
	topAdDomains        *prometheus.GaugeVec
	topSources          *prometheus.GaugeVec
	forwardDestinations *prometheus.GaugeVec
	queryTypes          *prometheus.GaugeVec
)

// DailyCounter is used to convert daily counts into monotonically
// increasing counters
type DailyCounter struct {
	prometheus.Counter
	Value float64
}

func (d *DailyCounter) GetIncrease(newValue float64) float64 {
	v := newValue

	if v-d.Value >= 0 {
		v -= d.Value
	}
	d.Value = newValue

	return v
}

func (d *DailyCounter) Update(newValue float64) {
	v := d.GetIncrease(newValue)
	d.Add(v)
}

func buildMetrics() *prometheus.Registry {
	registry := prometheus.NewRegistry()

	blockedDomains = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "blocked_domains",
		Namespace: namespace,
		Help:      "Total number of domains blocked by pi-hole",
	})

	dnsQueries = &DailyCounter{
		Counter: prometheus.NewCounter(prometheus.CounterOpts{
			Name:      "dns_queries",
			Namespace: namespace,
			Help:      "Total number of dns queries",
		}),
	}

	blockedAds = &DailyCounter{
		Counter: prometheus.NewCounter(prometheus.CounterOpts{
			Name:      "blocked_ads",
			Namespace: namespace,
			Help:      "Total number of blocked dns queries",
		}),
	}

	forwardedQueries = &DailyCounter{
		Counter: prometheus.NewCounter(prometheus.CounterOpts{
			Name:      "forwarded_queries",
			Namespace: namespace,
			Help:      "Total number of forwarded dns queries",
		}),
	}

	cachedQueries = &DailyCounter{
		Counter: prometheus.NewCounter(prometheus.CounterOpts{
			Name:      "cached_queries",
			Namespace: namespace,
			Help:      "Total number of dns query cache hits",
		}),
	}

	uniqueDomains = &DailyCounter{
		Counter: prometheus.NewCounter(prometheus.CounterOpts{
			Name:      "unique_domains",
			Namespace: namespace,
			Help:      "Total number of unique requested domains",
		}),
	}

	clients = &DailyCounter{
		Counter: prometheus.NewCounter(prometheus.CounterOpts{
			Name:      "clients",
			Namespace: namespace,
			Help:      "Total number of clients",
		}),
	}

	uniqueClients = &DailyCounter{
		Counter: prometheus.NewCounter(prometheus.CounterOpts{
			Name:      "unique_clients",
			Namespace: namespace,
			Help:      "Total number of unique clients",
		}),
	}

	replies = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "replies",
			Namespace: namespace,
			Help:      "Number of dns replies for a given type",
		},
		[]string{"type"},
	)

	topDomains = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "top_domains",
			Namespace: namespace,
			Help:      "Number of queries for today's top ten most queried domains",
		},
		[]string{"domain"},
	)

	topAdDomains = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "top_ad_domains",
			Namespace: namespace,
			Help:      "Number of queries for today's top ten most blocked domains",
		},
		[]string{"domain"},
	)

	topSources = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "top_sources",
			Namespace: namespace,
			Help:      "Number of queries from today's top ten most active clients",
		},
		[]string{"source"},
	)

	forwardDestinations = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "forward_destinations",
			Namespace: namespace,
			Help:      "Percentage of queries forwarded to a given destination",
		},
		[]string{"destination"},
	)

	queryTypes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "query_types",
			Namespace: namespace,
			Help:      "Percentage of queries by type",
		},
		[]string{"type"},
	)

	metrics := []prometheus.Collector{
		blockedDomains,
		dnsQueries,
		blockedAds,
		forwardedQueries,
		cachedQueries,
		uniqueDomains,
		clients,
		uniqueClients,
		replies,
		topDomains,
		topAdDomains,
		topSources,
		forwardDestinations,
		queryTypes,
	}

	for _, metric := range metrics {
		registry.MustRegister(metric)
	}

	return registry
}

func updateMetrics(piholeHost, token string) {
	stats, err := queryPihole(http.DefaultClient, piholeHost, token)
	if err != nil {
		log.Printf("Unable to query pihole API: %s", err)
		return
	}

	blockedDomains.Set(stats.DomainsBeingBlocked)
	dnsQueries.Update(stats.DNSQueriesToday)
	blockedAds.Update(stats.AdsBlockedToday)
	forwardedQueries.Update(stats.QueriesForwarded)
	cachedQueries.Update(stats.QueriesCached)
	uniqueDomains.Update(stats.UniqueDomains)
	clients.Update(stats.ClientsEverSeen)
	uniqueClients.Update(stats.UniqueClients)

	replies.WithLabelValues("nodata").Set(stats.ReplyNODATA)
	replies.WithLabelValues("nxdomain").Set(stats.ReplyNXDOMAIN)
	replies.WithLabelValues("cname").Set(stats.ReplyCNAME)
	replies.WithLabelValues("ip").Set(stats.ReplyIP)

	for domain, queries := range stats.TopQueries {
		topDomains.WithLabelValues(domain).Set(queries)
	}

	for domain, queries := range stats.TopAds {
		topAdDomains.WithLabelValues(domain).Set(queries)
	}

	for source, queries := range stats.TopSources {
		topSources.WithLabelValues(source).Set(queries)
	}

	for destination, percent := range stats.ForwardDestinations {
		forwardDestinations.WithLabelValues(destination).Set(percent)
	}

	for queryType, percent := range stats.QueryTypes {
		queryTypes.WithLabelValues(queryType).Set(percent)
	}
}
