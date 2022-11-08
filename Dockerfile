FROM golang:1.19-alpine
WORKDIR /workspace/
COPY . .

RUN apk update && \
	apk --no-cache add gcc g++ upx
RUN CGO_ENABLED=1 GOOS=linux go build -a --ldflags '-linkmode external -extldflags "-static"' -o pihole-exporter *.go

FROM scratch
WORKDIR /root/
COPY --from=0 /workspace/pihole-exporter /root/pihole-exporter
COPY --from=tarampampam/curl:latest /bin/curl /root/curl

EXPOSE 9617/tcp
HEALTHCHECK --interval=60s --timeout=2s --retries=3 --start-period=2s \
	CMD ["/root/curl", "--fail", "--url", "http://localhost:9617/readiness"]

CMD ["/root/pihole-exporter"]
