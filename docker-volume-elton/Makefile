TARGETDIR = ../bin
VOLUME = $(TARGETDIR)/docker-volume-elton

all: build

build:
	GOOS=linux GOARCH=amd64 godep go build -o $(VOLUME)

clean:
	$(RM) $(VOLUME)

.PHONY: all build clean
