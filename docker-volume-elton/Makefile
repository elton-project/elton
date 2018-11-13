TARGETDIR = ../bin
VOLUME = $(TARGETDIR)/docker-volume-elton

all: build

build-deps:
	which go

build:
	GOOS=linux GOARCH=amd64 go build -o $(VOLUME)

clean:
	$(RM) $(VOLUME)

.PHONY: all build clean
