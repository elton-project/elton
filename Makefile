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
	$(MAKE) -C server/elton

eltonfs: deps
	$(MAKE) -C eltonfs/eltonfs

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

.PHONY: all deps grpc client build install eltonfs test testall clean
