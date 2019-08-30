.PHONY: all
all: build-deps generate fmt test build

.PHONY: build-deps
build-deps:
	which go
	which docker
	which docker-compose
	which dep || go get github.com/golang/dep/cmd/dep
	$(MAKE) -C api build-deps

.PHONY: generate
generate:
	$(MAKE) -C api generate

.PHONY: build
build:
	go build ./cmd/eltond
	$(MAKE) -C eltonfs build-inside-container

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: test
test:
	go test ./...

.PHONY: clean
clean:
	$(MAKE) -C eltonfs clean
