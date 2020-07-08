BINARY=pihole-exporter
BUILD=$$(vtag --no-meta)

default: docker

build: format
	go build -o ${BINARY} -v ./cmd/notes

docker: format
	docker build --network=host -t ${BINARY}:${BUILD} -f Dockerfile .

format fmt:
	go fmt -x ./...

clean:
	go mod tidy
	go clean
	rm -f $(BINARY)

get-tag:
	echo ${BUILD}

.PHONY: all build format fmt clean get-tag
