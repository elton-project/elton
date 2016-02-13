TARGETDIR = ../../bin
FS = $(TARGETDIR)/eltonfs

all: build

build:
	godep go build -o $(FS)

clean:
	$(RM) $(FS)

.PHONY: all build clean
