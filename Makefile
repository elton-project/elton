TARGETDIR = bin

all: fmt grpc elton eltonfs volume-plugin

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
