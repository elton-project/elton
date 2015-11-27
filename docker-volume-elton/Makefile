TARGETDIR = ../bin
VOLUME = $(TARGETDIR)/docker-volume-elton

all: build

build:
	godep go build -o $(VOLUME)

clean:
	$(RM) $(VOLUME)

.PHONY: all build clean
