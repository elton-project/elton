TARGETDIR = bin
TARGET = $(TARGETDIR)/elton

all: deps build

$(TARGET):

deps:
	go get -d -v

build: $(TARGET)
	go build -o $^

install: $(TARGET)
	go install

clean:
	$(RM) -r $(TARGETDIR)

.PHONY: all $(TARGET) deps build install clean
