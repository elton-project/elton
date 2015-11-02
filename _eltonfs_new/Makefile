TARGETDIR = ../bin
FS = $(TARGETDIR)/eltonfs

all: install

$(FS):

deps:
	godep get

build: $(FS)
	godep go build -o $^

install: deps
	godep go install

test: deps
	godep go test

clean:
	$(RM) $(FS)

.PHONY: all $(FS) test deps build install clean
