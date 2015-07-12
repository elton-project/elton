TARGETDIR = ../bin
CLIENT = $(TARGETDIR)/elton

all: deps build

$(CLIENT):

deps:
	go get -d -v

build: $(CLIENT)
	go build -o $^

clean:
	$(RM) $(CLIENT)

.PHONY: all $(CLIENT) deps build clean
