TARGETDIR = ../../bin
FS = $(TARGETDIR)/eltonfs

all: build

build:
	godep go build -o $(FS)

install:
	godep go install

test:
	godep go test

clean:
	$(RM) $(FS)

.PHONY: all test build install clean
