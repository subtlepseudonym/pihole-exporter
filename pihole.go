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
	forward
;`

type PiholeStats struct {
	QueryTypes          map[string]float64
	AllowedQueries      map[string]map[string]float64
	BlockedQueries      map[string]float64
	BlockedCNAMEQueries map[string]float64
	ClientQueries       map[string]float64
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
	"gravity",         // during deep CNAME inspection
	"regex_blacklist", // during deep CNAME inspection
	"exact_blacklist", // during deep CNAME inspection
	"retried_query",
	"retried_ignored_query",
	"already_forwarded",
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
			numQueries float64
			client     string
			forward    sql.NullString
		)

		err = rows.Scan(&queryType, &status, &client, &forward, &numQueries)
		if err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		if queryType >= len(queryTypes) {
			return nil, fmt.Errorf("unknown query type: %d", queryType)
		}

		typeKey := queryTypes[queryType-1]
		stats.QueryTypes[typeKey] += numQueries

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
		case 1, 4, 5, 6, 7, 8:
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
