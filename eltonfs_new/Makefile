TARGETDIR = ../bin
FS = $(TARGETDIR)/eltonfs

all: deps build

$(FS):

deps:
	go get -d -v

build: $(FS)
	go build -o $^

install:
	go install

test: deps
	go test

clean:
	$(RM) $(FS)

.PHONY: all $(FS) test deps build install clean
