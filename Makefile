TARGETDIR = bin

all: deps client eltonfs

binary:
	docker build -t eltonbuilder .
	docker run --rm -it --privileged -v $(CURDIR):/vendor/src/git.t-lab.cs.teu.ac.jp/nashio/elton eltonbuilder

deps:
	godep get ./...

grpc:
	$(MAKE) -C grpc/proto

client: deps
	$(MAKE) -C cmd

eltonfs: deps
	$(MAKE) -C eltonfs

build:
	docker build -f cmd/Dockerfile -t elton .
#	docker build -f eltonfs/Dockerfile -t eltonfs .

test:
	$(MAKE) -C api test
	$(MAKE) -C grpc test
	$(MAKE) -C eltonfs test

testall: build
	docker-compose up

clean:
	$(RM) -r $(TARGETDIR)

.PHONY: all deps grpc client build install eltonfs test testall clean
