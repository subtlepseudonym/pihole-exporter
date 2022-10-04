BINARY=pihole-exporter

BUILD=$$( \
	if command -v vtag &>/dev/null; then \
		vtag --no-meta; \
	else \
		printf \
			'0.0.1-unknown+%s' \
			"$$(git rev-list -n1 HEAD | head -c7)"; \
	fi \
)
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
