.PHONY: all
all: build-deps generate

.PHONY: build-deps
build-deps:
	which go
	which protoc-gen-go || go get -v github.com/golang/protobuf/protoc-gen-go

.PHONY: generate
generate:
	protoc --go_out=plugins=grpc:. *.proto
