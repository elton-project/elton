#include <linux/module.h>
#include <linux/kernel.h>

static int __init fs_module_init(void) {
    printk(KERN_INFO "sample module loaded\n");
    return 0;
}

static void __exit fs_module_exit(void) {
    printk(KERN_INFO "sample module unloaded\n");
}

module_init(fs_module_init);
module_exit(fs_module_exit);

MODULE_LICENSE("GPL v2");
MODULE_AUTHOR("yuuki0xff <yuuki0xff@gmail.com>");
MODULE_DESCRIPTION("sample module");
