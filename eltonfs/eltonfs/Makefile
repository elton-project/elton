TARGETDIR = ../../bin
FS = $(TARGETDIR)/eltonfs

all: build

build:
	GOOS=linux GOARCH=amd64 go build -o $(FS)

clean:
	$(RM) $(FS)

.PHONY: all build clean
