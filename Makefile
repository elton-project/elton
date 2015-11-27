TARGETDIR = bin

all: deps elton eltonfs volume-plugin

binary:
	docker build -t eltonbuilder .
	docker run --rm -it --privileged -v $(CURDIR):/vendor/src/git.t-lab.cs.teu.ac.jp/nashio/elton eltonbuilder

deps:
	godep get ./...

grpc:
	$(MAKE) -C grpc/proto

elton: deps
	$(MAKE) -C server/elton

eltonfs: deps
	$(MAKE) -C eltonfs/eltonfs

volume-plugin: deps
	$(MAKE) -C docker-volume-elton

build:
	docker build -f server/elton/Dockerfile -t elton .
#	docker build -f eltonfs/Dockerfile -t eltonfs .

test:
	$(MAKE) -C server/elton test
	$(MAKE) -C grpc/proto test
	$(MAKE) -C eltonfs/eltonfs test

testall: build
	docker-compose up

clean:
	$(RM) -r $(TARGETDIR)

.PHONY: all deps grpc elton build install eltonfs volume-plugin test testall clean
