GODEPS = \
	github.com/codegangsta/cli \
	github.com/boltdb/bolt \
	github.com/bmizerany/pat \
	github.com/fukata/golang-stats-api-handler

BIN = elton

all: install

.deps:
	go get -u $(GODEPS)

build: .deps
	go build

install: .deps
	go install

clean:
	rm -rf $(GOPATH)/bin/$(BIN) $(BIN)
