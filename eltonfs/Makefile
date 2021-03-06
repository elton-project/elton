MODULE_NAME := elton
FS_NAME := elton
LTP_DIR := /mnt/ltp
CC_FLAGS += -g
CC_FLAGS += -I $(PWD)/include
# CC_FLAGS += -DELTONFS_DEBUG
CC_FLAGS += -DELTONFS_STATISTIC
CC_FLAGS += -DELTONFS_XATTRS
# CC_FLAGS += -DELTONFS_UNIT_TEST
# CC_FLAGS += -DELTON_RPC_DEBUG
# CC_FLAGS += -DELTON_RPC_CALL_TEST
CC := gcc $(CC_FLAGS)

elton-y += super.o
elton-y += simple_xattr.o xattr.o
elton-y += assert.o
elton-y += commit.o
elton-y += inode.o
elton-y += dir.o
elton-y += file.o
elton-y += symlink.o
elton-y += local_cache.o
elton-y += ioctl.o
elton-y += xdr/default.o \
	xdr/bin_encoding.o \
	xdr/test.o
elton-y += rpc/test.o \
	rpc/packet.o \
	rpc/queue.o \
	rpc/server.o \
	rpc/server_handler.o \
	rpc/server_io.o \
	rpc/server_session.o \
	rpc/server_ns.o
obj-m := elton.o

KERNEL ?= $(shell uname -r)

# Need following packages:
#   * linux-headers-amd64
KDIR ?= /lib/modules/$(KERNEL)/build

DOCKER_IMAGE := eltonfs-builder
DOCKER_RUN = docker run --rm -it \
	-u $(shell id -u):$(shell id -g) \
	-v $(abspath .):/work \
	-w /work \
	$(DOCKER_IMAGE)

.PHONY: all
all: generate build test

.PHONY: generate
generate:
	$(MAKE) -C clustertest generate

.PHONY: build-inside-container
build-inside-container:
	docker build -t $(DOCKER_IMAGE) builder
	$(DOCKER_RUN) make build

.PHONY: clean-inside-container
clean-inside-container:
	docker build -t $(DOCKER_IMAGE) builder
	$(DOCKER_RUN) make clean

.PHONY: build
build:
	$(MAKE) -C $(KDIR) M=`pwd` modules

.PHONY: clean
clean:
	$(MAKE) -C $(KDIR) M=`pwd` clean

.PHONY: test
test: test_basic_op test_syscalls test_libc

.PHONY: mount
mount: mount-umh mount-kmod

.PHONY: mount-umh
mount-umh:
	# Start eltond and prepare a volume before mounting.
	ELTON_ROLES=controller,storage setsid --fork eltond </dev/null >eltond.log 2>&1 &
	sleep 1

.PHONY: mount-kmod
mount-kmod:
	elton volume create foo
	elton volume ls |grep foo
	mount -t $(FS_NAME) -o vol=foo dummy /mnt

.PHONY: test_basic_op
test_basic_op:
	# Test basic operations.
	touch /mnt/a
	ls /mnt
	ls -l /mnt
	stat /mnt/a
	chmod a=rw /mnt/a
	echo foo >/mnt/a
	grep foo /mnt/a
	rm /mnt/a
	mount |grep /mnt |grep ' type $(MODULE_NAME) (rw,relatime,.*)$$'

.PHONY: test_syscalls
test_syscalls:
	$(MAKE) -C test/syscalls

.PHONY: test_libc
test_libc:
	$(MAKE) -C test/libc

.PHONY: test_ltp
test_ltp:
	mkdir -p $(LTP_DIR)
	ulimit -n `ulimit -Hn` && \
	if /opt/ltp/runltp -s mmap -d $(LTP_DIR) -S $(LTP_DIR)/skipfile -o $(LTP_DIR)/outfile; then \
		echo OK; \
	else \
		sed -n '/^tag=/p; /^Summary:$$/,+4p' $(LTP_DIR)/outfile |grep -E -A2 -B5 '^failed   [^0]'; \
	fi

.PHONY: rmmod
rmmod:
	umount /mnt
	rmmod $(MODULE_NAME)

.PHONY: force_rmmod
force_rmmod:
	# umount
	umount /mnt || :
	! grep /proc/mounts /mnt
	# unload the module.
	rmmod $(MODULE_NAME) || :
	! grep $(MODULE_NAME) /proc/modules
