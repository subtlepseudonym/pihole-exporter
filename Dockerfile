FROM golang:1.14-alpine
WORKDIR /workspace/
COPY . .

RUN apk update && \
	apk --no-cache add gcc g++ upx
RUN CGO_ENABLED=1 GOOS=linux go build -a --ldflags '-linkmode external -extldflags "-static"' -o pihole-exporter *.go

# compress binary
RUN upx -f --brute pihole-exporter

FROM scratch
WORKDIR /root/
COPY --from=0 /workspace/pihole-exporter /root/pihole-exporter

CMD ["/root/pihole-exporter"]
