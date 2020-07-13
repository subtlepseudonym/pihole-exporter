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
	count(id)
FROM
	queries
WHERE
	timestamp >= %d AND
	timestamp < %d
GROUP BY
	type,
	status,
	client
;`

type PiholeStats struct {
	QueryTypes          map[string]float64
	AllowedQueries      map[string]float64
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
}

func queryPihole(db *sql.DB, since, now int64) (*PiholeStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultDBTimeout)
	defer cancel()

	stats := &PiholeStats{
		QueryTypes:          make(map[string]float64),
		AllowedQueries:      make(map[string]float64),
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
		var queryType, status int
		var numQueries float64
		var client string

		err = rows.Scan(&queryType, &status, &client, &numQueries)
		if err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		typeKey := queryTypes[queryType-1]
		stats.QueryTypes[typeKey] += numQueries
		stats.ClientQueries[client] += numQueries

		statusKey := queryStatuses[status]
		switch status {
		case 0, 2, 3:
			stats.AllowedQueries[statusKey] += numQueries
		case 1, 4, 5, 6, 7, 8:
			stats.BlockedQueries[statusKey] += numQueries
		case 9, 10, 11:
			stats.BlockedCNAMEQueries[statusKey] += numQueries
		default:
			return nil, fmt.Errorf("unexpected status: %d", status)
		}
	}

	return stats, nil
}
