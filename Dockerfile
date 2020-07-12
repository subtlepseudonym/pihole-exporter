FROM golang:1.14-alpine
WORKDIR /workspace/
COPY . .

RUN apk update && \
	apk --no-cache add upx
RUN CGO_ENABLED=0 GOOS=linux go build -a -o pihole-exporter *.go
RUN upx -f --brute pihole-exporter

FROM scratch
WORKDIR /root/
COPY --from=0 /workspace/pihole-exporter /root/pihole-exporter

CMD ["/root/pihole-exporter"]
