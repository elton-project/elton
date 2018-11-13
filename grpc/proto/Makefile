all: deps generate

deps:
	# TODO: rename rule name to 'build-deps'
	go get -v github.com/golang/protobuf/protoc-gen-go

generate:
	protoc --go_out=plugins=grpc:. elton_service.proto

.PHONY: all deps generate	
