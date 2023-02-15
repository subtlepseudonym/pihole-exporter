## Pihole Exporter

[![github](https://img.shields.io/github/v/tag/subtlepseudonym/pihole-exporter?logo=github)](https://github.com/subtlepseudonym/pihole-exporter) [![docker_pulls](https://img.shields.io/docker/pulls/subtlepseudonym/pihole-exporter?logo=docker)](https://hub.docker.com/r/subtlepseudonym/pihole-exporter)  [![kofi](https://img.shields.io/badge/ko--fi-Support%20me%20-hotpink?logo=kofi&logoColor=white)](https://ko-fi.com/subtlepseudonym)

### Running the container

```bash
docker create \
	--name pihole-exporter \
	--env "PIHOLE_DSN=file:/pihole/pihole-FTL.db?_query_only" \
	--volume "/local/path/to/ftl.db:/pihole/pihole-FTL.db:ro" \
	subtlepseudonym/pihole-exporter:latest
```

### Building this project

This project is intended to be run in docker.
```bash
docker build \
	--tag subtlepseudonym/pihole-exporter:latest \
	--file ./Dockerfile \
	.
```
Alternatively, you can run `make docker` if you've cloned the repo. This will automatically
tag the image as well.

NOTE: you should have `vtag` from [subtlepseudonym/utilities](https://github.com/subtlepseudonym/utilities) installed for auto-tagging to work properly. Without it, docker version tag will be `0.0.1-unknown`

### Metrics

| Metric | Description |
| ------ | ----------- |
| pihole_dns_queries_total | Total number of DNS queries with type labels |
| pihole_client_dns_queries | Total number of DNS queries with client labels |
| pihole_allowed_dns_queries | Forwarded or cached DNS queries |
| pihole_blocked_dns_queries | Blocked DNS queries |
| pihole_query_replies | DNS query replies with reply type labels |
| pihole_exporter_http_request_duration_seconds | How long this exporter takes to respond on the `/metrics` endpoint |

### Motivation
As of creating this project, the top two hits on google for `pihole exporter` use only gauges
to represent the data scraped from Pihole. It has been expressed that scraped metrics
[should be instrumented from the point of view of the thing being instrumented](https://github.com/prometheus-net/prometheus-net/issues/63#issuecomment-360070401),
implying that the use of gauges is the correct way to instrument the pihole API. Despite this,
I hewed closer to the [first-party prometheus instrumentation guidelines](https://prometheus.io/docs/practices/instrumentation/) and sought to provide
metrics whose type better represented the nature of the value being measured.

Specifically,
there are many values exposed by the pihole API that perform daily counts. These daily counts
are generated with a rolling stepwise function such that all queries (or blocked ads, clients, etc)
in the last 23 hours plus those since the last hour are counted and the value is updated every hour.
In practice, this leads to values that decrease by the amount of queries received during the upcoming
one hour block, yesterday. I believe that this doesn't do a very good job of representing what should
be monotonically increasing values; you can't un-make a DNS request. To retrieve absolute counts, I
chose to make requests against the FTL database rather than the pihole API. This makes requests to
the `/metrics` endpoint take a bit longer and require that this exporter has access to the pihole
database file (which could be a prohibitive requirement depending upon your setup), but I believe
it provides better metrics.
