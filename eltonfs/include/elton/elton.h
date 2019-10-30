#ifndef _ELTON_ELTON_H
#define _ELTON_ELTON_H

#include <linux/fs.h>
#include <linux/kernel.h>
#include <linux/xattr.h>

#define MODULE_NAME "elton"
#define FS_NAME "elton"
#define ELTONFS_SUPER_MAGIC 0x51891f5
#define ELTONFS_NAME_LEN 255
// Path to executable file of helper process.
#define ELTONFS_HELPER "eltonfs-helper"
// Path to UNIX domain socket to communicate with helper process.
#define ELTONFS_HELPER_SOCK "/run/eltonfs-helper.sock"
// PATH Environment value for elton-helper.
#define PATH_ENV "/sbin:/usr/sbin:/bin:/usr/bin"

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
  if (inode == NULL) {
    return NULL;
  }
  return &(inode->vfs_inode);
}
static inline struct eltonfs_inode *eltonfs_i(struct inode *inode) {
  if (inode == NULL) {
    return NULL;
  }
  return container_of(inode, struct eltonfs_inode, vfs_inode);
}

#endif // _ELTON_ELTON_H
