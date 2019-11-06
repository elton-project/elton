PREFIX = /usr/local
MODULE_DIR = /lib/modules/$(shell uname -r)/elton

BUILD_SBIN_FILES += build/sbin/eltond
BUILD_SBIN_FILES += build/sbin/eltonfs-helper
BUILD_KMOD_FILES += build/kmod/elton.ko
BUILD_FILES += $(BUILD_SBIN_FILES) $(BUILD_KMOD_FILES)

GO_DEPS = Makefile go.* $(shell find */ -name '*.go')
KMOD_DEPS = Makefile \
	$(shell git ls-files |grep '^eltonfs/' ) \
	$(shell find eltonfs/ -name '*.c' -o -name '*.h')



.PHONY: all
all: build-deps generate fmt build-dev test-fast

.PHONY: build-deps
build-deps:
	which go
	which docker
	which docker-compose
	which dep || go get github.com/golang/dep/cmd/dep
	$(MAKE) -C api build-deps

.PHONY: generate
generate:
	$(MAKE) -C api generate

.PHONY: build
build: $(BUILD_FILES)

.PHONY: build-dev
build-dev:
	__ELTON_BUILD_IN_CONTAINER=1 $(MAKE) build

.PHONY: install
install:
	install -D -m 700 $(BUILD_SBIN_FILES) $(PREFIX)/sbin/
	install -D -m 644 $(BUILD_KMOD_FILES) $(MODULE_DIR)/
	depmod

.PHONY: fmt
fmt:
	go fmt ./...
	find ./api/ -name '*.proto' |xargs -r clang-format -i
	find ./eltonfs/ -name '*.c' -o -name '*.h' |xargs -r clang-format -i

.PHONY: test-fast
test-fast:
	go test -cover -timeout 5s ./...
	$(MAKE) -C eltonfs/clustertest test-fast

.PHONY: test
test: test-fast

.PHONY: clean
clean:
	rm -rf build/
	$(MAKE) -C eltonfs clean



build/sbin/eltond: $(GO_DEPS)
	go build -o $@ ./cmd/eltond

build/sbin/eltonfs-helper: $(GO_DEPS)
	go build -o $@ ./cmd/eltonfs-helper

build/kmod/elton.ko: eltonfs/elton.ko
	install -D -m 644 $< $@

eltonfs/elton.ko: $(KMOD_DEPS)
	if [ -z "$(__ELTON_BUILD_IN_CONTAINER)" ]; then \
		$(MAKE) -C eltonfs build; \
	else \
		$(MAKE) -C eltonfs build-inside-container; \
    fi
