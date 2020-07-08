FROM golang:1.14
WORKDIR /workspace/
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o pihole-exporter *.go

FROM scratch
WORKDIR /root/
COPY --from=0 /workspace/pihole-exporter /root/pihole-exporter

ARG pihole_api_token
ENV PIHOLE_API_TOKEN=${pihole_api_token}

CMD ["/root/pihole-exporter"]
