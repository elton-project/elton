TARGETDIR = ../../bin
CLIENT = $(TARGETDIR)/elton

all: build

build-deps:
	which go

build:
	GOOS=linux GOARCH=amd64 go build -o $(CLIENT)

clean:
	$(RM) $(CLIENT)

.PHONY: all $(CLIENT) build clean
