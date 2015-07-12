TARGETDIR = ../bin
FS = $(TARGETDIR)/eltonfs

all: deps build

$(FS):

deps:
	go get -d -v

build: $(FS)
	go build -o $^

clean:
	$(RM) $(FS)

.PHONY: all $(FS) deps build clean
