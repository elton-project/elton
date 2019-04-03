#include <linux/module.h>
#include <linux/kernel.h>
#include <linux/fs.h>
#include <linux/dcache.h>

#define MODULE_NAME "simple_fs"
#define FS_NAME MODULE_NAME

#define _PRINTLNK(level, fmt, ...) (printk(level MODULE_NAME ": " fmt "\n", ##__VA_ARGS__))
#define DEBUG(fmt, ...) _PRINTLNK(KERN_DEBUG, fmt, ##__VA_ARGS__)
#define INFO(fmt, ...) _PRINTLNK(KERN_INFO, fmt, ##__VA_ARGS__)
#define ERR(fmt, ...) _PRINTLNK(KERN_ERR, fmt, ##__VA_ARGS__)

// エラーならtrueを返す。
// また、エラー発生時にログを残す。
#define CHECK_ERROR(error) ({ \
	if(error) { \
		ERR("Occurred an error %d on %s (%s:%d)", error, __func__, __FILE__, __LINE__); \
	} \
	error; \
})



static int simplefs_fill_super(struct super_block *sb, void *data, int silent) {
	// not implemented
	return -ENOSYS;
}
static struct dentry *mount(struct file_system_type *fs_type,
		int flags, const char *dev_name, void *data) {
	return mount_nodev(fs_type, flags, data, simplefs_fill_super);
}
static void kill_sb(struct super_block *sb) {}


static struct file_system_type simplefs_type = {
	.name = "simple_fs",
	.mount = mount,
	.kill_sb = kill_sb,
	.fs_flags = 0
};

static int __init fs_module_init(void) {
	int error;
	DEBUG("Loading the module ...");

	error = register_filesystem(&simplefs_type);
	if(CHECK_ERROR(error)) {
		return error;
	}

	INFO("The module loaded");
	return 0;
}

static void __exit fs_module_exit(void) {
	int error;
	DEBUG("Unloading the module ...");

	error = unregister_filesystem(&simplefs_type);
	if(CHECK_ERROR(error)) {
		return;
	}

	INFO("The module unloaded");
}

module_init(fs_module_init);
module_exit(fs_module_exit);

MODULE_LICENSE("GPL v2");
MODULE_AUTHOR("yuuki0xff <yuuki0xff@gmail.com>");
MODULE_DESCRIPTION(MODULE_NAME " module");
