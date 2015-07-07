TARGETDIR = bin
CLIENT = $(TARGETDIR)/elton
FS = $(TARGETDIR)/eltonfs
CONFIGCMD = mkconfig
CLIOBJS = cmd/main.go cmd/version.go cmd/commands.go
FSOBJS = eltonfs/main.go eltonfs/fs.go eltonfs/node.go eltonfs/files.go

all: deps build

$(CLIENT): $(CLIOBJS)
	go build -o $@ $^

$(FS): $(FSOBJS)
	go build -o $@ $^

deps:
	go get -d -v

build: $(CLIENT) $(FS)

clean:
	$(RM) -r $(TARGETDIR)

.PHONY: all $(CLIENT) $(FS) deps build clean
