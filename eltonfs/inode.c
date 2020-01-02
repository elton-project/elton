// Inode Number Flavors
// ====================
//
// vfs_ino:
//     Inode number used in the Linux VFS.  It is commonly called "ino".  This
//     number must unique per super block.  We can access it from indoe->i_ino.
// eltonfs_ino:
//     Inode number used in the Eltonfs.  This number must unique per eltonfs
//     volume.  We can access it from eltonfs_i(inode)->eltonfs_ino.  It is
//     assigned by elton controllers.
//
//
//
// VFS Inode Number Assignation Rules
// ==================================
//
// The rule to be applied differ depending on inode status.  See the table and
// rule descriptions below.
//
//                                  | RuleA    RuleB    RuleC    RuleD
//     -----------------------------+---------------------------------
//     Is eltonfs_ino assigned?     |   N        Y        N        Y
//     Is local_cache_id assigned?  |   N        N        Y        Y
//
// RuleA:
//     This situation occurs when creating new file or directory.  Generate any
//     number between ELTONFS_LOCAL_INO_MIN and ELTONFS_LOCAL_INO_MAX, assign it
//     to the inode.  Make sure duplication checking of vfs_ino.
// RuleB:
//     This situation occurs when accessing an committed files.  Assign
//     eltonfs_ino to vfs_ino.
// RuleC:
//     This situation will occurs when accessing an not-committed files.  This
//     inode assigned vfs_ino according to the RuleA before it is evicted.
//     Should use previous vfs_ino assigned by RuleA.
// RuleD:
//     If created some commits without unmount, some inodes are assigned two
//     different ino (eltonfs_ino and vfs_ino).  So those are fall into a
//     inconsistent state.  We have to solve with following tricks until
//     unmount.
//     * lookup() (called from getdents) should emit with vfs_ino.
//     * iget() should search and build an inode with vfs_ino if search key are
//       within LOCAL_INO range.
#include <elton/elton.h>
#include <elton/rpc/struct.h>
#include <elton/utils.h>
#include <linux/pagemap.h>

#define ELTONFS_LOCAL_INO_MIN (U64_MAX ^ (u64)U32_MAX)
#define ELTONFS_LOCAL_INO_MAX U64_MAX

// Initialize inode->i_op and inode->i_fop and inode->i_mapping.
// Should set inode->i_mode before call it.
void eltonfs_inode_init_ops(struct inode *inode, dev_t dev) {
  // todo: change aops by file types.
  inode->i_mapping->a_ops = &eltonfs_aops;
  mapping_set_gfp_mask(inode->i_mapping, GFP_HIGHUSER);
  // TODO: inodeのデータを永続化に対応してから、evictableにする。
  mapping_set_unevictable(inode->i_mapping);

  switch (inode->i_mode & S_IFMT) {
  default:
    init_special_inode(inode, inode->i_mode, dev);
    break;
  case S_IFREG:
    inode->i_op = &eltonfs_file_inode_operations;
    inode->i_fop = &eltonfs_file_operations;
    break;
  case S_IFDIR:
    inode->i_op = &eltonfs_dir_inode_operations;
    inode->i_fop = &eltonfs_dir_operations;

    /* directory inodes start off with i_nlink == 2 (for "." entry) */
    inc_nlink(inode);
    break;
  case S_IFLNK:
    inode->i_op = &eltonfs_symlink_inode_operations;
    inode_nohighmem(inode);
    break;
  }
}
// Initialize eltonfs internal data.
void eltonfs_inode_init_internal(struct inode *inode) {
  struct eltonfs_inode *ei = eltonfs_i(inode);
  ei->eltonfs_ino = 0;
  spin_lock_init(&ei->lock);
  simple_xattrs_init(&ei->xattrs);

  switch (inode->i_mode & S_IFMT) {
  case S_IFREG:
    ei->file.object_id = NULL;
    ei->file.local_cache_id = NULL;
    ei->file.cache_inode = NULL;
    break;
  case S_IFDIR:
    INIT_LIST_HEAD(&ei->dir.dir_entries._list_head);
    ei->dir.count = 0;
    break;
  case S_IFLNK:
    ei->symlink.object_id = NULL;
    ei->symlink.redirect_to = NULL;
    break;
  }
}
void eltonfs_inode_init_regular(struct inode *inode, const char *object_id,
                                const char *local_cache_id) {
  struct eltonfs_inode *ei = eltonfs_i(inode);
  ei->file.object_id = dup_string_direct(object_id);
  ei->file.local_cache_id = dup_string_direct(local_cache_id);
  ei->file.cache_inode = NULL;
  // todo: error check
}
void eltonfs_inode_init_dir(struct inode *inode) {
  struct eltonfs_inode *ei = eltonfs_i(inode);
  INIT_LIST_HEAD(&ei->dir.dir_entries._list_head);
  ei->dir.count = 0;
}
void eltonfs_inode_init_symlink(struct inode *inode, const char *object_id) {
  struct eltonfs_inode *ei = eltonfs_i(inode);
  ei->symlink.object_id = dup_string_direct(object_id);
  ei->symlink.redirect_to = NULL;
  // todo: error check
}

// Get inode from backend tree by eltonfs_ino.
struct eltonfs_inode *eltonfs_iget(struct super_block *sb, u64 ino) {
  struct eltonfs_info *info = eltonfs_sb(sb);
  struct eltonfs_inode_xdr *i_xdr;
  struct inode *inode;
  struct eltonfs_inode *ei;

  i_xdr = radix_tree_lookup(info->inodes_ei, ino);
  if (!i_xdr)
    return NULL;

  inode = new_inode(sb);
  if (!inode)
    return ERR_PTR(-ENOMEM);
  ei = eltonfs_i(inode);

  WARN_ONCE(i_xdr->eltonfs_ino != ino, "ino is not match");
  ei->eltonfs_ino = ino;
  inode->i_ino = ino;
  inode->i_mode = i_xdr->mode;
  inode->i_uid.val = i_xdr->owner;
  inode->i_gid.val = i_xdr->group;
  inode->i_atime = timestamp_to_timespec64(i_xdr->atime);
  inode->i_mtime = timestamp_to_timespec64(i_xdr->mtime);
  inode->i_ctime = timestamp_to_timespec64(i_xdr->ctime);
  inode->i_rdev = MKDEV(i_xdr->major, i_xdr->minor);

  eltonfs_inode_init_ops(inode, inode->i_rdev);
  eltonfs_inode_init_internal(inode);

  switch (inode->i_mode & S_IFMT) {
  case S_IFREG:
    eltonfs_inode_init_regular(inode, i_xdr->object_id, NULL);
    break;
  case S_IFDIR: {
    int error;
    eltonfs_inode_init_dir(inode);
    error = dup_dir_entries(&ei->dir.dir_entries, &i_xdr->dir_entries);
    if (error)
      return ERR_PTR(error);
    ei->dir.count = i_xdr->dir_entries_len;
    break;
  }
  case S_IFLNK:
    eltonfs_inode_init_symlink(inode, i_xdr->object_id);
    break;
  }
  return eltonfs_i(inode);
}

// Generate new vfs_ino.
u64 eltonfs_get_next_ino(struct super_block *sb) {
  struct eltonfs_info *info = eltonfs_sb(sb);
  u64 ino;
retry:
  ino = ++info->last_local_ino;
  if (unlikely(ino < ELTONFS_LOCAL_INO_MIN)) {
    info->last_local_ino = ELTONFS_LOCAL_INO_MIN;
    goto retry;
  }
  if (unlikely(radix_tree_lookup(info->inodes_vfs, ino)))
    goto retry;
  return ino;
}

// Create inode under specified directory.
// The content of created inode is only stored only local storage until commit
// operation is executed.
struct inode *eltonfs_create_inode(struct super_block *sb,
                                   const struct inode *dir, umode_t mode,
                                   dev_t dev) {
  struct inode *inode;
  inode = new_inode(sb);
  if (!inode) {
    return inode;
  }

  inode->i_ino = eltonfs_get_next_ino(sb);
  inode_init_owner(inode, dir, mode);
  inode->i_atime = inode->i_mtime = inode->i_ctime = current_time(inode);
  eltonfs_inode_init_ops(inode, dev);
  eltonfs_inode_init_internal(inode);
  return inode;
}
