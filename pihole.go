package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

const defaultDBTimeout = 30 * time.Second

const piholeQuery = `
SELECT
	type,
	status,
	client,
	forward,
	reply_type,
	count(id)
FROM
	queries
WHERE
	timestamp >= %d AND
	timestamp < %d
GROUP BY
	type,
	status,
	client,
	forward,
	reply_type
;`

type PiholeStats struct {
	QueryTypes          map[string]float64
	AllowedQueries      map[string]map[string]float64
	BlockedQueries      map[string]float64
	BlockedCNAMEQueries map[string]float64
	ClientQueries       map[string]float64
	QueryReplies        map[string]float64
}

var queryTypes = []string{
	"A",
	"AAAA",
	"ANY",
	"SRV",
	"SOA",
	"PTR",
	"TXT",
	"NAPTR",
	"MX",
	"DS",
	"RRSIG",
	"DNSKEY",
	"NS",
	"OTHER",
	"SVCB",
	"HTTPS",
}

var queryStatuses = []string{
	"unknown",
	"gravity",
	"forwarded",
	"cache_hit",
	"regex_blacklist",
	"exact_blacklist",
	"known_upstream",
	"unspecified_upstream",
	"nxdomain_upstream",
	"gravity_cname",         // during deep CNAME inspection
	"regex_blacklist_cname", // during deep CNAME inspection
	"exact_blacklist_cname", // during deep CNAME inspection
	"retried_query",
	"retried_ignored_query",
	"already_forwarded",
	"database_busy",
	"special_domain",
}

var replyTypes = []string{
	"unknown",
	"nodata",
	"nxdomain",
	"cname",
	"ip",
	"domain",
	"rrname",
	"servfail",
	"refused",
	"notimp",
	"other",
	"dnssec",
	"none", // query was dropped intentionally
	"blob", // binary data
}

func queryPihole(db *sql.DB, since, now int64) (*PiholeStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultDBTimeout)
	defer cancel()

	stats := &PiholeStats{
		QueryTypes:          make(map[string]float64),
		AllowedQueries:      make(map[string]map[string]float64),
		BlockedQueries:      make(map[string]float64),
		BlockedCNAMEQueries: make(map[string]float64),
		ClientQueries:       make(map[string]float64),
		QueryReplies:        make(map[string]float64),
	}

	query := fmt.Sprintf(piholeQuery, since, now)
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query db: %w", err)
	}
	for rows.Next() {
		var (
			queryType  int
			status     int
			client     string
			forward    sql.NullString
			replyType  int
			numQueries float64
		)

		err = rows.Scan(&queryType, &status, &client, &forward, &replyType, &numQueries)
		if err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		if queryType < 1 || queryType > len(queryTypes) {
			return nil, fmt.Errorf("unknown query type: %d", queryType)
		}

		typeKey := queryTypes[queryType-1]
		stats.QueryTypes[typeKey] += numQueries

		if replyType < 1 || replyType > len(replyTypes) {
			return nil, fmt.Errorf("unknown reply type: %d", replyType)
		}

		replyKey := replyTypes[replyType-1]
		stats.QueryReplies[replyKey] += numQueries

		stats.ClientQueries[client] += numQueries

		switch status {
		case 0, 2, 3, 12, 13, 14:
			statusKey := queryStatuses[status]
			upstream := "cache"
			if forward.Valid || status == 0 {
				upstream = forward.String
			}
			if stats.AllowedQueries[statusKey] == nil {
				stats.AllowedQueries[statusKey] = make(map[string]float64)
			}
			stats.AllowedQueries[statusKey][upstream] += numQueries
		case 1, 4, 5, 6, 7, 8, 15, 16:
			statusKey := queryStatuses[status]
			stats.BlockedQueries[statusKey] += numQueries
		case 9, 10, 11:
			statusKey := queryStatuses[status]
			stats.BlockedCNAMEQueries[statusKey] += numQueries
		default:
			return nil, fmt.Errorf("unexpected status: %d", status)
		}
	}

	return stats, nil
}
