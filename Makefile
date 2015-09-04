TARGETDIR = bin

all: client eltonfs

binary:
	@docker build -t eltonbuilder .
	@docker run --rm -it --privileged -v $(CURDIR)/bin:/elton/bin eltonbuilder

client:
	@$(MAKE) -C cmd

eltonfs:
	@$(MAKE) -C eltonfs

build:
	@docker build -f cmd/Dockerfile -t elton .
	@docker build -f eltonfs/Dockerfile -t eltonfs .

test:
	@$(MAKE) -C api test
	@$(MAKE) -C grpc test
	@$(MAKE) -C eltonfs test

testall: test build
	@docker run --rm -it elton /bin/bash

clean:
	@$(RM) -r $(TARGETDIR)

.PHONY: all client build install eltonfs test testall clean
