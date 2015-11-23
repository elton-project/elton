TARGETDIR = ../bin
CLIENT = $(TARGETDIR)/elton

all: build

$(CLIENT):

build: $(CLIENT)
	godep go build -o $^

install:
	godep go install

clean:
	$(RM) $(CLIENT)

.PHONY: all $(CLIENT) build install clean
