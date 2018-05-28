TARGETDIR = bin

all: deps fmt grpc elton eltonfs volume-plugin

binary:
	docker build -t eltonbuilder .
	docker run --rm -it --privileged -v $(CURDIR):/vendor/src/gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton eltonbuilder

deps:
	godep restore -v

fmt:
	go fmt ./...

grpc: deps
	$(MAKE) -C grpc/proto

elton: deps
	$(MAKE) -C server/elton

eltonfs: deps
	$(MAKE) -C eltonfs/eltonfs

volume-plugin: deps
	$(MAKE) -C docker-volume-elton

build:
	docker build -f server/elton/Dockerfile -t elton .

testall: build
	docker-compose up

clean:
	$(RM) -r $(TARGETDIR)

.PHONY: all deps fmt grpc elton eltonfs volume-plugin build testall clean
