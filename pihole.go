package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	defaultScheme  = "http"
	piholeEndpoint = "admin/api.php"
)

// piholeOptions define the response from the pihole API
// https://discourse.pi-hole.net/t/pi-hole-api/1863
var piholeOptions = []string{
	"summaryRaw",
	"overTimeData",
	"topItems",
	"recentItems",
	"getQueryTypes",
	"getForwardDestinations",
	"getQuerySources",
	"jsonForceObject",
}

type PiholeResponse struct {
	DomainsBeingBlocked float64            `json:"domains_being_blocked"`
	DNSQueriesToday     float64            `json:"dns_queries_today"`
	AdsBlockedToday     float64            `json:"ads_blocked_today"`
	UniqueDomains       float64            `json:"unique_domains"`
	QueriesForwarded    float64            `json:"queries_forwarded"`
	QueriesCached       float64            `json:"queries_cached"`
	ClientsEverSeen     float64            `json:"clients_ever_seen"`
	UniqueClients       float64            `json:"unique_clients"`
	DNSQueriesAllTypes  float64            `json:"dns_queries_all_types"`
	ReplyNODATA         float64            `json:"reply_NODATA"`
	ReplyNXDOMAIN       float64            `json:"reply_NXDOMAIN"`
	ReplyCNAME          float64            `json:"reply_CNAME"`
	ReplyIP             float64            `json:"reply_IP"`
	TopQueries          map[string]float64 `json:"top_queries"`
	TopAds              map[string]float64 `json:"top_ads"`
	TopSources          map[string]float64 `json:"top_sources"`
	ForwardDestinations map[string]float64 `json:"forward_destinations"`
	QueryTypes          map[string]float64 `json:"querytypes"`
}

func queryPihole(client *http.Client, host, apiToken string) (*PiholeResponse, error) {
	url := fmt.Sprintf("%s://%s/%s?%s&auth=%s", defaultScheme, host, piholeEndpoint, strings.Join(piholeOptions, "&"), apiToken)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("new pihole request: %w", err)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do pihole request: %w", err)
	}

	if res.Body == nil {
		return nil, fmt.Errorf("nil pihole response body")
	}

	var stats PiholeResponse
	err = json.NewDecoder(res.Body).Decode(&stats)
	if err != nil {
		return nil, fmt.Errorf("decode pihole response: %w", err)
	}

	return &stats, nil
}
