GODEPS = \
	github.com/codegangsta/cli \
	github.com/boltdb/bolt \
	github.com/bmizerany/pat \
	github.com/fukata/golang-stats-api-handler \
	github.com/golang/glog

all: install

.deps:
	go get -u $(GODEPS)

build: .deps
	go build

install: .deps
	go install

clean:
	go clean
