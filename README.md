## Pihole Exporter

### Running this project

This project is intended to be run in docker.
```bash
docker build \
	--tag subtlepseudonym/pihole-exporter:latest \
	--file ./Dockerfile \
	.
```
Note: if you're using pihole, you may want to set `--network=host`

```bash
docker create \
	--name pihole-exporter \
	--env "PIHOLE_HOST=local.pihole.address" \
	--env "PIHOLE_API_TOKEN=very_secret" \
	subtlepseudonym/pihole-exporter:latest
```

### Motivation
As of creating this project, the top two hits on google for `pihole exporter` use only gauges
to represent the data scraped from Pihole. While it has been expressed that scraped metrics
[should be instrumented from the point of view of the thing being instrumented](https://github.com/prometheus-net/prometheus-net/issues/63#issuecomment-360070401),
I prefer something closer to the [prometheus instrumentation guidelines](https://prometheus.io/docs/practices/instrumentation/). Specifically,
there are many values exposed by the pihole API that perform daily counts. Rather than using
a gauge to represent these, I wanted to use a counter.
