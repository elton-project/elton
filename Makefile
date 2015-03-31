GODEPS = \
	github.com/codegangsta/cli \
	github.com/fukata/golang-stats-api-handler \
	github.com/BurntSushi/toml \
	github.com/go-sql-driver/mysql

TARGETDIR = bin
TARGET = $(TARGETDIR)/elton

all: deps build

$(TARGET):

deps:
	go get -u $(GODEPS)

build: $(TARGET)
	go build -o $^

install: $(TARGET)
	go install

clean:
	$(RM) -r $(TARGETDIR)

.PHONY: all $(TARGET) deps build install clean
