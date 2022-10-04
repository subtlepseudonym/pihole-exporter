BINARY=pihole-exporter
BUILD=$$(([[ ! -z "$$(which vtag)" ]] && vtag --no-meta) || echo '0.0.1-unknown')
TAG="${BINARY}:${BUILD}"

default: test docker

build: format
	go build -o ${BINARY} -v .

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
