TARGETDIR = ../../bin
CLIENT = $(TARGETDIR)/elton

all: build

build:
	GOOS=linux GOARCH=amd64 godep go build -o $(CLIENT)

clean:
	$(RM) $(CLIENT)

.PHONY: all $(CLIENT) build clean
