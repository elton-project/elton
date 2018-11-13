TARGETDIR = bin

all: build-deps fmt grpc elton eltonfs volume-plugin

build-deps:
	which go
	which docker
	which docker-compose
	$(MAKE) -C docker-volume-elton build-deps
	$(MAKE) -C eltonfs/eltonfs     build-deps
	$(MAKE) -C grpc/proto          build-deps
	$(MAKE) -C server/elton        build-deps

binary:
	docker build -t eltonbuilder .
	docker run --rm -it --privileged -v $(CURDIR):/vendor/src/gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton eltonbuilder

fmt:
	go fmt ./...

grpc:
	$(MAKE) -C grpc/proto

elton:
	$(MAKE) -C server/elton

eltonfs:
	$(MAKE) -C eltonfs/eltonfs

volume-plugin:
	$(MAKE) -C docker-volume-elton

build:
	docker build -f server/elton/Dockerfile -t elton .

testall: build
	docker-compose up

clean:
	$(RM) -r $(TARGETDIR)

.PHONY: all fmt grpc elton eltonfs volume-plugin build testall clean
