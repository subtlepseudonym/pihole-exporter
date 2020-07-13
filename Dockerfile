FROM golang:1.14-alpine
WORKDIR /workspace/
COPY . .

RUN apk update && \
	apk --no-cache add gcc g++
RUN CGO_ENABLED=1 GOOS=linux go build -a --ldflags '-linkmode external -extldflags "-static"' -o pihole-exporter *.go

FROM scratch
WORKDIR /root/
COPY --from=0 /workspace/pihole-exporter /root/pihole-exporter

CMD ["/root/pihole-exporter"]
