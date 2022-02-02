BINARY=pihole-exporter
BUILD=$$(vtag --no-meta)
TAG="${BINARY}:${BUILD}"

default: test docker

build: format
	go build -o ${BINARY} -v ./cmd/notes

docker: format
	docker build --network=host -t ${TAG} -f Dockerfile .

test: format
	gotest --race ./...

format fmt:
	go fmt -x ./...

clean:
	go mod tidy
	go clean
	rm -f $(BINARY)

get-tag:
	echo ${BUILD}

.PHONY: all build format fmt clean get-tag
