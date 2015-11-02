TARGETDIR = bin

all: client eltonfs

binary:
	docker build -t eltonbuilder .
	docker run --rm -it --privileged -v $(CURDIR):/vendor/src/git.t-lab.cs.teu.ac.jp/nashio/elton eltonbuilder

client:
	$(MAKE) -C cmd

eltonfs:
	$(MAKE) -C eltonfs

build:
	docker build -f cmd/Dockerfile -t elton .
#	docker build -f eltonfs/Dockerfile -t eltonfs .
#	docker build -f munin/Dockerfile -t munin munin

test:
	$(MAKE) -C api test
	$(MAKE) -C grpc test
	$(MAKE) -C eltonfs test

testall: build
	docker-compose up

clean:
	$(RM) -r $(TARGETDIR)

.PHONY: all client build install eltonfs test testall clean
