#include <elton/elton.h>
#include <elton/rpc/struct.h>
#include <elton/utils.h>
#include <linux/pagemap.h>

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

// Get inode from backend tree by eltonfs_ino.
struct eltonfs_inode *eltonfs_iget(struct super_block *sb, u64 ino) {
  struct eltonfs_info *info = eltonfs_sb(sb);
  struct eltonfs_inode_xdr *i_xdr;
  struct inode *inode;
  struct eltonfs_inode *ei;

  i_xdr = radix_tree_lookup(info->cinfo->tree->inodes, ino);
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

  switch (inode->i_mode & S_IFMT) {
  case S_IFREG: {
    char *oid;
    int error = dup_string(&oid, i_xdr->object_id);
    if (error)
      return ERR_PTR(error);
    ei->file.object_id = oid;
    ei->file.local_cache_id = NULL;
    ei->file.cache_inode = NULL;
    break;
  }
  case S_IFDIR: {
    int error = dup_dir_entries(&ei->dir.dir_entries, &i_xdr->dir_entries);
    if (error)
      return ERR_PTR(error);
    ei->dir.count = i_xdr->dir_entries_len;
    break;
  }
  case S_IFLNK: {
    int error;
    char *oid;
    error = dup_string(&oid, i_xdr->object_id);
    if (error)
      return ERR_PTR(error);
    ei->symlink.object_id = oid;
    ei->symlink.redirect_to = NULL;
    break;
  }
  }
  return eltonfs_i(inode);
}
