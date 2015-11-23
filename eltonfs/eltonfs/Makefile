TARGETDIR = ../../bin
FS = $(TARGETDIR)/eltonfs

all: build

$(FS):

build: $(FS)
	godep go build -o $^

install:
	godep go install

test:
	godep go test

clean:
	$(RM) $(FS)

.PHONY: all $(FS) test build install clean
