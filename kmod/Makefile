MODULE_NAME := simple_fs
FS_NAME := simple_fs
LTP_DIR := /mnt/ltp

obj-m := $(MODULE_NAME).o

# Need following packages:
#   * linux-headers-amd64
KDIR := /lib/modules/$(shell uname -r)/build


.PHONY: all
all: build test

.PHONY: build
build:
	$(MAKE) -C $(KDIR) M=$(PWD) modules

.PHONY: test
test: force_rmmod insmod test_basic_op test_syscalls rmmod

.PHONY: insmod
insmod:
	insmod $(MODULE_NAME).ko
	# Check if the module is loaded.
	grep $(MODULE_NAME) /proc/modules
	# Check if the fs is mountable.
	mount -t $(FS_NAME) dummy /mnt

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
	mount |grep /mnt |fgrep ' type simple_fs (rw,relatime)'

.PHONY: test_syscalls
test_syscalls:
	$(MAKE) -C test/syscalls

.PHONY: test_ltp
test_ltp:
	mkdir -p $(LTP_DIR)
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
