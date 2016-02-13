TARGETDIR = ../../bin
CLIENT = $(TARGETDIR)/elton

all: build

build:
	godep go build -o $(CLIENT)

clean:
	$(RM) $(CLIENT)

.PHONY: all $(CLIENT) build clean
