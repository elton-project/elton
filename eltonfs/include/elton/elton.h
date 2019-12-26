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
#define ELTONFS_HELPER "/usr/local/sbin/eltonfs-helper"
// Path to UNIX domain socket to communicate with helper process.
#define ELTONFS_HELPER_SOCK "/run/elton.sock"
#define ELTONFS_HELPER_OUTPUT "/var/log/eltonfs-helper.log"
#define ELTONFS_HELPER_LOG_TAG "eltonfs-helper-output"
// PATH Environment value for elton-helper.
#define PATH_ENV                                                               \
  "/sbin:"                                                                     \
  "/usr/sbin:"                                                                 \
  "/usr/local/sbin:"                                                           \
  "/bin:"                                                                      \
  "/usr/bin:"                                                                  \
  "/usr/local/bin"

struct eltonfs_config {
  // Flag for auto_tx mode  (default=true).
  // If this flag is true, transaction should start in following situations:
  //  * eltonfs is mounted.
  //  * transaction is rollbacked.
  bool auto_tx;
  // Volume id  (nullable)
  char *vid;
  // Commit id  (nullable)
  char *cid;
  // Volume name  (nullable)
  char *vol_name;
};

// FS specified data.  It is linked from super block.
struct eltonfs_info {
  struct eltonfs_config config;

#ifdef ELTONFS_STATISTIC
  unsigned long mmap_size;
  rwlock_t mmap_size_lock;
#endif
};

struct eltonfs_dir_entry {
  struct list_head _list_head;
  // Eltonfs inode number.
  u64 ino;
  // Length of file name.
  u8 name_len;
  // File name.
  // NOTE: increase array size to append NULL at the end of string.
  char name[ELTONFS_NAME_LEN + 1];
};

struct eltonfs_dir_entry_ino {
  struct list_head _list_head;
  char file[ELTONFS_NAME_LEN];
  // Inode number
  u64 ino;
};

struct eltonfs_inode {
  struct inode vfs_inode;

  // Inode number for internal use.
  u64 eltonfs_ino;

  spinlock_t lock;
  union {
    struct {
      // Remote object ID.
      // This ID must assigned by elton-storage.  If this inode created by this
      // node and it is not committed, object_id must NULL.
      const char *object_id;
      // Local object ID.
      // This ID must only use this node.
      const char *local_cache_id;
      // Inode of the cache file.
      struct inode *cache_inode;
    } file;

    struct {
      // List of entries contained in this directory.
      // TODO: store directory entries to persistent storage.
      struct eltonfs_dir_entry dir_entries;
      // Number of dir entries.
      u64 count;
    } dir;

    // Linked list of eltonfs_dir_entry_ino.
    // This field using only while initializing directory tree.
    struct eltonfs_dir_entry_ino *_dir_entries_tmp;

    struct {
      // Remote object ID.
      // See eltonfs_inode.file.object_id for detail.
      const char *object_id;
      // Redirect path.
      // If redirect_to is NULL, try to get an object from elton-storage.
      const char *redirect_to;
    } symlink;
  };

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

void eltonfs_inode_init_once(struct eltonfs_inode *i);

// Iterate all directory entries in a eltonfs_inode.
//
// Arguments:
//   i:         int or size_t
//   entry_ino: struct eltonfs_dir_entry_ino *
#define ELTONFS_FOR_EACH_DIRENT(eltonfs_inode, entry)                          \
  if (S_ISREG((eltonfs_inode)->vfs_inode.i_mode)) {                            \
    ERR("try to iterate directory entries of non-directory inode");            \
    BUG();                                                                     \
  } else                                                                       \
    list_for_each_entry((entry), &(eltonfs_inode)->dir.dir_entries._list_head, \
                        _list_head)

#endif // _ELTON_ELTON_H
