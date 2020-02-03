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

extern struct elton_rpc_server server;

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

  // Current CommitID.
  const char *cid;
  // Current commit information  (MUST NOT USE)
  // MUST only use to keep pointer to commit info.  If unmounting or changed
  // commit, we should release memory by elton_rpc_free_decoded_data() and
  // inodes_ei should replace with other tree.  The inode tree
  // (cinfo->tree->inodes) and entries are modified by eltonfs.
  struct commit_info *cinfo;
  // Credentials of mount time.
  // Should release by put_cred() when unmounting.
  const struct cred *cred;

  // Inode tree (lookup by eltonfs_ino).
  // Alias of cinfo->tree->inodes.
  //
  // Key: eltonfs_ino
  // Value: struct eltonfs_inode_xdr *
  struct radix_tree_root *inodes_ei;
  // Inode tree (lookup by vfs_ino).
  //
  // Key: vfs_ino
  // Value: struct eltonfs_inode_xdr *
  struct radix_tree_root *inodes_vfs;
  // Last used local VFS ino.
  // If you need new vfs_ino, use eltonfs_get_next_ino() instead of directly
  // access it.
  u64 last_local_ino;
  spinlock_t last_local_ino_lock;
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

struct eltonfs_inode {
  struct inode vfs_inode;

  // Inode number for internal use.
  u64 eltonfs_ino;

  // This spinlock should use when reading or updating of file specified fields.
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
      // Type of elements: struct eltonfs_dir_entry *
      // TODO: store directory entries to persistent storage.
      struct list_head dir_entries;
      // Number of dir entries.
      u64 count;
    } dir;

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

extern const struct file_operations eltonfs_file_operations;
extern const struct inode_operations eltonfs_file_inode_operations;
extern const struct file_operations eltonfs_dir_operations;
extern const struct inode_operations eltonfs_dir_inode_operations;
extern const struct inode_operations eltonfs_symlink_inode_operations;
long eltonfs_ioctl(struct file *file, unsigned int cmd, unsigned long arg);
long eltonfs_compat_ioctl(struct file *file, unsigned int cmd,
                          unsigned long arg);
int elton_update_time(struct inode *inode, struct timespec64 *time, int flags);
struct inode *eltonfs_create_inode(struct super_block *sb,
                                   const struct inode *dir, umode_t mode,
                                   dev_t dev);
struct eltonfs_inode *eltonfs_iget(struct super_block *sb, u64 ino);

static inline struct inode *vfs_i(struct eltonfs_inode *inode) {
  if (inode == NULL || IS_ERR(inode))
    return ERR_CAST(inode);
  return &(inode->vfs_inode);
}
static inline struct eltonfs_inode *eltonfs_i(struct inode *inode) {
  if (inode == NULL || IS_ERR(inode))
    return ERR_CAST(inode);
  return container_of(inode, struct eltonfs_inode, vfs_inode);
}

static inline struct eltonfs_info *eltonfs_sb(struct super_block *sb) {
  if (sb == NULL || IS_ERR(sb))
    return ERR_CAST(sb);
  return sb->s_fs_info;
}

void eltonfs_inode_init_once(struct eltonfs_inode *i);
void eltonfs_inode_init_ops(struct inode *inode, dev_t dev);
void eltonfs_inode_init_regular(struct inode *inode, const char *object_id,
                                const char *local_cache_id);
int eltonfs_inode_init_regular_with_new_cache(struct inode *inode);
void eltonfs_inode_init_dir(struct inode *inode);
void eltonfs_inode_init_symlink(struct inode *inode, const char *object_id,
                                const char *redirect_to);
struct inode *eltonfs_get_obj_inode(const char *oid, struct super_block *sb);

// Iterate all directory entries in a eltonfs_inode.
// Should acquire lock of the directory before calling it.
//
// Arguments:
//   eltonfs_inode: Inode of directory.
//   entry:         struct eltonfs_dir_entry *
#define ELTONFS_FOR_EACH_DIRENT(eltonfs_inode, entry)                          \
  if (S_ISREG((eltonfs_inode)->vfs_inode.i_mode)) {                            \
    ERR("try to iterate directory entries of non-directory inode");            \
    BUG();                                                                     \
  } else                                                                       \
    list_for_each_entry((entry), &(eltonfs_inode)->dir.dir_entries, _list_head)

// ioctl commands
#define ELTONFS_IOCTL_MAGIC 183
#define ELTONFS_IOC_BEGIN _IOR(ELTONFS_IOCTL_MAGIC, 1, char __user *)
#define ELTONFS_IOC_COMMIT _IOW(ELTONFS_IOCTL_MAGIC, 2, char __user *)
#define ELTONFS_IOC_ROLLBACK _IOW(ELTONFS_IOCTL_MAGIC, 3, char __user *)

#endif // _ELTON_ELTON_H
