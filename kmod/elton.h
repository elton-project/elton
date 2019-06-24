#ifndef _ELTON_ELTON_H
#define _ELTON_ELTON_H

#include <linux/kernel.h>
#include <linux/fs.h>
#include <linux/xattr.h>


#define MODULE_NAME "elton"
#define FS_NAME "elton"
#define ELTONFS_SUPER_MAGIC 0x51891f5
#define ELTONFS_NAME_LEN 255

struct eltonfs_info {
#ifdef ELTONFS_STATISTIC
	unsigned long mmap_size;
	rwlock_t mmap_size_lock;
#endif
};

struct eltonfs_inode {
	struct inode vfs_inode;

#ifdef ELTONFS_XATTRS
	struct simple_xattrs xattrs;
#endif
};


static inline struct inode *vfs_i(struct eltonfs_inode *inode) {
	if(inode == NULL) {
		return NULL;
	}
	return &(inode->vfs_inode);
}
static inline struct eltonfs_inode *eltonfs_i(struct inode *inode) {
	if(inode == NULL) {
		return NULL;
	}
	return container_of(inode, struct eltonfs_inode, vfs_inode);
}

#endif // _ELTON_ELTON_H
