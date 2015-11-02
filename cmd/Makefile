TARGETDIR = ../bin
CLIENT = $(TARGETDIR)/elton

all: deps build

$(CLIENT):

deps:
	godep get

build: $(CLIENT)
	go build -a -o $^

install:
	go install

clean:
	$(RM) $(CLIENT)

.PHONY: all $(CLIENT) deps build install clean
