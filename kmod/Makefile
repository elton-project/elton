MODULE_NAME := simple_fs
FS_NAME := simple_fs

obj-m := $(MODULE_NAME).o

# Need following packages:
#   * linux-headers-amd64
KDIR := /lib/modules/$(shell uname -r)/build


all: build test

build:
	$(MAKE) -C $(KDIR) M=$(PWD) modules

test: force_rmmod insmod test_fs rmmod

insmod:
	insmod $(MODULE_NAME).ko

test_fs:
	# Check if the module is loaded.
	grep $(MODULE_NAME) /proc/modules
	# Check if the fs is mountable.
	mount -t $(FS_NAME) dummy /mnt

rmmod:
	rmmod $(MODULE_NAME)

force_rmmod:
	rmmod $(MODULE_NAME) || :

