.PHONY: all
all: build-deps generate fmt build test-fast

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
	go build ./cmd/eltonfs-helper
	$(MAKE) -C eltonfs build-inside-container

.PHONY: fmt
fmt:
	go fmt ./...
	find ./api/ -name '*.proto' |xargs -r clang-format -i
	find ./eltonfs/ -name '*.c' -o -name '*.h' |xargs -r clang-format -i

.PHONY: test-fast
test-fast:
	go test -cover -timeout 5s ./...
	$(MAKE) -C eltonfs/clustertest test-fast

.PHONY: test
test: test-fast

.PHONY: clean
clean:
	$(MAKE) -C eltonfs clean
