TARGETDIR = bin
TARGET = $(TARGETDIR)/elton
CONFIGCMD = mkconfig

all: deps build

$(TARGET):

deps:
	go get -d -v

build: $(TARGET)
	go build -o $^

config:
	$(CONFIGCMD) examples/config.tml config.tml

backupconfig:
	$(CONFIGCMD) examples/backup.tml config.tml

install: $(TARGET)
	go install

clean:
	$(RM) -r $(TARGETDIR)

.PHONY: all $(TARGET) deps build config backupconfig install clean
