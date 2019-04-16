#include <linux/module.h>
#include <linux/kernel.h>
#include <linux/fs.h>
#include <linux/dcache.h>
#include <linux/pagemap.h>

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
#define ASSERT_NOT_NULL(p) ({ \
	if(!p) { \
		ERR(#p " is NULL "); \
		BUG_ON(p); \
	} \
	p; \
})



static bool is_registered = 0;
static struct file_system_type simplefs_type;
static struct super_operations simplefs_s_op;
static struct address_space_operations simplefs_aops;
static struct inode_operations simplefs_file_inode_operations;
static struct inode_operations simplefs_dir_inode_operations;
static struct file_operations simplefs_file_operations;


struct inode *simplefs_get_inode(struct super_block *sb,
				const struct inode *dir, umode_t mode, dev_t dev) {
	struct inode *inode;
	inode = new_inode(sb);
	if(! inode) {
		return inode;
	}

	inode->i_ino = get_next_ino();
	inode_init_owner(inode, dir, mode);
	inode->i_mapping->a_ops = &simplefs_aops;
	mapping_set_gfp_mask(inode->i_mapping, GFP_HIGHUSER);
	mapping_set_unevictable(inode->i_mapping);
	inode->i_atime = inode->i_mtime = inode->i_ctime = current_time(inode);
	switch (mode & S_IFMT) {
	default:
		init_special_inode(inode, mode, dev);
		break;
	case S_IFREG:
		inode->i_op = &simplefs_file_inode_operations;
		inode->i_fop = &simplefs_file_operations;
		break;
	case S_IFDIR:
		inode->i_op = &simplefs_dir_inode_operations;
		inode->i_fop = &simple_dir_operations;

		/* directory inodes start off with i_nlink == 2 (for "." entry) */
		inc_nlink(inode);
		break;
	case S_IFLNK:
		inode->i_op = &page_symlink_inode_operations;
		inode_nohighmem(inode);
		break;
	}
	return inode;
}

static int simplefs_set_page_dirty(struct page *page) {
	if(PageDirty(page)) {
		return 0;
	}
	SetPageDirty(page);
	return 0;
}

int simplefs_mknod(struct inode *dir, struct dentry *dentry, umode_t mode, dev_t dev) {
	struct inode *inode = simplefs_get_inode(dir->i_sb, dir, mode, dev);
	if(! inode) {
		return -ENOSPC;
	}
	d_instantiate(dentry, inode);
	dget(dentry);
	dir->i_mtime = dir->i_ctime = current_time(dir);
	return 0;
}

int simplefs_create(struct inode *dir, struct dentry *dentry, umode_t mode, bool excl) {
	return simplefs_mknod(dir, dentry, mode | S_IFREG, 0);
}

int simplefs_mkdir(struct inode *dir, struct dentry *dentry, umode_t mode) {
	int error = simplefs_mknod(dir, dentry, mode | S_IFDIR, 0);
	if(error) {
		return error;
	}
	inc_nlink(dir);
	return 0;
}

int simplefs_symlink(struct inode *dir, struct dentry *dentry, const char *symname) {
	struct inode *inode;
	int len, error;

	inode = simplefs_get_inode(dir->i_sb, dir, S_IFLNK | S_IRWXUGO, 0);
	if(! inode) {
		return -ENOSPC;
	}
	len = strlen(symname) + 1;
	// TODO: allocate physical pages.
	error = page_symlink(inode, symname, len);
	if(error){
		iput(inode);
	}
	d_instantiate(dentry, inode);
	dget(dentry);
	dir->i_mtime = dir->i_ctime = current_time(dir);
	return 0;
}

static unsigned long simplefs_get_unmapped_area(struct file *file, unsigned long addr, unsigned long len, unsigned long pgoff, unsigned long flags) {
	return current->mm->get_unmapped_area(file, addr, len, pgoff, flags);
}

static int simplefs_fill_super(struct super_block *sb, void *data, int silent) {
	struct inode *inode;
	struct dentry *root;

	DEBUG("Preparing for super block ...");
	save_mount_options(sb, data);
	sb->s_blocksize_bits = PAGE_SHIFT;
	sb->s_blocksize = PAGE_SIZE;
	sb->s_maxbytes = PAGE_SIZE;
	sb->s_type = &simplefs_type;
	sb->s_op = &simplefs_s_op;
	sb->s_time_gran = 1;

	inode = simplefs_get_inode(sb, NULL, S_IFDIR, 0);
	ASSERT_NOT_NULL(inode);
	root = d_make_root(inode);
	ASSERT_NOT_NULL(root);
	sb->s_root = root;
	DEBUG("Prepared the super block");
	return 0;
}
static struct dentry *mount(struct file_system_type *fs_type,
		int flags, const char *dev_name, void *data) {
	return mount_nodev(fs_type, flags, data, simplefs_fill_super);
}
static void kill_sb(struct super_block *sb) {}


static int __init fs_module_init(void) {
	int error;
	DEBUG("Loading the module ...");

	error = register_filesystem(&simplefs_type);
	if(CHECK_ERROR(error)) {
		return error;
	}

	is_registered = 1;
	INFO("The module loaded");
	return 0;
}

static void __exit fs_module_exit(void) {
	int error;
	DEBUG("Unloading the module ...");

	if(is_registered) {
		error = unregister_filesystem(&simplefs_type);
		if(CHECK_ERROR(error)) {
			return;
		}
	}

	INFO("The module unloaded");
}



static struct file_system_type simplefs_type = {
	.name = FS_NAME,
	.mount = mount,
	.kill_sb = kill_sb,
	.fs_flags = 0
};
static struct super_operations simplefs_s_op = {
	.statfs		= simple_statfs,
	.drop_inode	= generic_delete_inode,
	.show_options	= generic_show_options,
};
static struct address_space_operations simplefs_aops = {
	.readpage	= simple_readpage,
	.write_begin	= simple_write_begin,
	.write_end	= simple_write_end,
	.set_page_dirty	= simplefs_set_page_dirty,
};
static struct inode_operations simplefs_file_inode_operations = {
	.setattr = simple_setattr,
	.getattr = simple_getattr,
};
static struct inode_operations simplefs_dir_inode_operations = {
	.create = simplefs_create,
	.lookup = simple_lookup,
	.link = simple_link,
	.unlink = simple_unlink,
	.symlink = simplefs_symlink,
	.mkdir = simplefs_mkdir,
	.rmdir = simple_rmdir,
	.mknod = simplefs_mknod,
	.rename = simple_rename,
};
static struct file_operations simplefs_file_operations = {
	.read_iter = generic_file_read_iter,
	.write_iter = generic_file_write_iter,
	.mmap = generic_file_mmap,
	.fsync = noop_fsync,
	.splice_read = generic_file_splice_read,
	.splice_write = iter_file_splice_write,
	.llseek = generic_file_llseek,
	.get_unmapped_area = simplefs_get_unmapped_area,
};



module_init(fs_module_init);
module_exit(fs_module_exit);

MODULE_LICENSE("GPL v2");
MODULE_AUTHOR("yuuki0xff <yuuki0xff@gmail.com>");
MODULE_DESCRIPTION(MODULE_NAME " module");
