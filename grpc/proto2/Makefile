all: deps generate

build-deps:
	which go
	go get -v github.com/golang/protobuf/protoc-gen-go

generate:
	protoc --go_out=plugins=grpc:. *.proto

.PHONY: all deps generate
