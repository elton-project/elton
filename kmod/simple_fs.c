#include <linux/slab.h>
#include <linux/mount.h>
#include <linux/module.h>
#include <linux/kernel.h>
#include <linux/fs.h>
#include <linux/dcache.h>
#include <linux/pagemap.h>
#include <linux/seq_file.h>
#include <linux/statfs.h>

#define MODULE_NAME "simple_fs"
#define FS_NAME MODULE_NAME
#define SIMPLEFS_SUPER_MAGIC 0x51891f5

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

struct simplefs_info {
#ifdef SIMPLEFS_STATISTIC
	unsigned long mmap_size;
	rwlock_t mmap_size_lock;
#endif
};


static struct inode *simplefs_get_inode(struct super_block *sb,
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

static int simplefs_mknod(struct inode *dir, struct dentry *dentry, umode_t mode, dev_t dev) {
	struct inode *inode = simplefs_get_inode(dir->i_sb, dir, mode, dev);
	if(! inode) {
		return -ENOSPC;
	}
	d_instantiate(dentry, inode);
	dget(dentry);
	dir->i_mtime = dir->i_ctime = current_time(dir);
	return 0;
}

static int simplefs_create(struct inode *dir, struct dentry *dentry, umode_t mode, bool excl) {
	return simplefs_mknod(dir, dentry, mode | S_IFREG, 0);
}

static int simplefs_mkdir(struct inode *dir, struct dentry *dentry, umode_t mode) {
	int error = simplefs_mknod(dir, dentry, mode | S_IFDIR, 0);
	if(error) {
		return error;
	}
	inc_nlink(dir);
	return 0;
}

static int simplefs_symlink(struct inode *dir, struct dentry *dentry, const char *symname) {
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

static long simplefs_ioctl(struct file *file, unsigned int cmd, unsigned long arg) {
	struct inode *inode = file_inode(file);
	unsigned int flags;

	switch(cmd) {
		case FS_IOC_GETFLAGS: {
			// TODO: 拡張属性に対応する。
			flags = 0;
			return put_user(flags, (int __user*)arg);
		}
		case FS_IOC_GETVERSION:
			return put_user(inode->i_generation, (int __user*)arg);
	}
	return -ENOSYS;  // Not implemented
}

#ifdef CONFIG_COMPAT
static long simplefs_compat_ioctl(struct file *file, unsigned int cmd, unsigned long arg) {
	switch(cmd) {
		case FS_IOC32_GETFLAGS:
			cmd = FS_IOC_GETFLAGS;
			break;
		case FS_IOC32_GETVERSION:
			cmd = FS_IOC_GETVERSION;
			break;
		default:
			return -ENOSYS;
	}
	return simplefs_ioctl(file, cmd, arg);
}
#endif

static int simplefs_fill_super(struct super_block *sb, void *data, int silent) {
	struct inode *inode;
	struct dentry *root;

	struct simplefs_info *info = kmalloc(sizeof(struct simplefs_info), GFP_KERNEL);
#ifdef SIMPLEFS_STATISTIC
	rwlock_init(&info->mmap_size_lock);
	info->mmap_size = 0;
#endif

	DEBUG("Preparing for super block ...");
	sb->s_blocksize_bits = PAGE_SHIFT;
	sb->s_blocksize = PAGE_SIZE;
	sb->s_maxbytes = MAX_LFS_FILESIZE;
	sb->s_type = &simplefs_type;
	sb->s_op = &simplefs_s_op;
	sb->s_time_gran = 1;
	sb->s_fs_info = info;

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

static int simplefs_statfs(struct dentry *dentry, struct kstatfs *buf) {
	// TODO: ダミーデータではなく、本当の値を設定する。
	int total_blocks = 10000;
	int used_blocks = 1000;
	int total_files = 1000;
	int used_files = 50;

	struct kstatfs stat = {
		/* Type of filesystem */
		.f_type = SIMPLEFS_SUPER_MAGIC,
		/* Optimal transfer block size */
		.f_bsize = PAGE_SIZE,
		/* Total data blocks in filesystem */
		.f_blocks = total_blocks,
		/* Free blocks in filesystem */
		.f_bfree = total_blocks - used_blocks,
		/* Free blocks available to unprivileged user */
		.f_bavail = total_blocks - used_blocks,
		/* Total file nodes in filesystem */
		.f_files = total_files,
		/* Free file nodes in filesystem */
		.f_ffree = total_files - used_files,
	};
	*buf = stat;
	return 0;
}

// Display the mount options in /proc/mounts.
static int simplefs_show_options(struct seq_file *m, struct dentry *root) {
	// seq_puts(m, ",default");
	return 0;
}


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

int simplefs_file_mmap(struct file *file, struct vm_area_struct *vma) {
#ifdef SIMPLEFS_STATISTIC
	struct simplefs_info *info = file->f_path.mnt->mnt_sb->s_fs_info;
	unsigned long size = vma->vm_end - vma->vm_start;
	int need_logging = 0;

	write_lock(&info->mmap_size_lock);
	if(info->mmap_size < size) {
		info->mmap_size = size;
		need_logging = 1;
	}
	write_unlock(&info->mmap_size_lock);

	if(need_logging)
		DEBUG("mmap size: file=%s, size=%ld", file->f_path.dentry->d_name.name, size);
#endif

	return generic_file_mmap(file, vma);
}


static struct file_system_type simplefs_type = {
	.owner = THIS_MODULE,
	.name = FS_NAME,
	.mount = mount,
	.kill_sb = kill_sb,
	.fs_flags = 0
};
static struct super_operations simplefs_s_op = {
	.statfs		= simplefs_statfs,
	.drop_inode	= generic_delete_inode,
	.show_options	= simplefs_show_options,
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
	.mmap = simplefs_file_mmap,
	.fsync = noop_fsync,
	.splice_read = generic_file_splice_read,
	.splice_write = iter_file_splice_write,
	.llseek = generic_file_llseek,
	.get_unmapped_area = simplefs_get_unmapped_area,
	.unlocked_ioctl = simplefs_ioctl,
#ifdef CONFIG_COMPAT
	.compat_ioctl = simplefs_compat_ioctl,  // for 32bit application.  See https://qiita.com/akachochin/items/94ba679b2941f55c1d2d
#endif
};



module_init(fs_module_init);
module_exit(fs_module_exit);

MODULE_ALIAS_FS("simple");
MODULE_LICENSE("GPL v2");
MODULE_AUTHOR("yuuki0xff <yuuki0xff@gmail.com>");
MODULE_DESCRIPTION(MODULE_NAME " module");
