all: deps generate

deps:
	godep get -v github.com/golang/protobuf/protoc-gen-go

generate:
	protoc --go_out=plugins=grpc:. elton_service.proto

.PHONY: all deps generate	
