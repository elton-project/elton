#include <elton/elton.h>
#include <elton/rpc/struct.h>
#include <linux/pagemap.h>

struct eltonfs_inode *eltonfs_iget(struct super_block *sb, u64 ino) {
  struct eltonfs_info *info = eltonfs_sb(sb);
  struct eltonfs_inode_xdr *i_xdr;
  struct inode *inode;

  i_xdr = radix_tree_lookup(info->cinfo->tree->inodes, ino);
  if (!i_xdr)
    return NULL;

  inode = new_inode(sb);
  if (!inode)
    return NULL;

  WARN_ONCE(i_xdr->eltonfs_ino != ino, "ino is not match");
  eltonfs_i(inode)->eltonfs_ino = ino;
  inode->i_ino = ino;
  inode->i_mode = i_xdr->mode;
  inode->i_uid.val = i_xdr->owner;
  inode->i_gid.val = i_xdr->group;
  // todo: convert from timestamp to timespec64.
  inode->i_atime;
  inode->i_mtime;
  inode->i_ctime;
  inode->i_rdev = MKDEV(i_xdr->major, i_xdr->minor);

  // todo: change aops by file types.
  inode->i_mapping->a_ops = &eltonfs_aops;
  mapping_set_gfp_mask(inode->i_mapping, GFP_HIGHUSER);
  // TODO: inodeのデータを永続化に対応してから、evictableにする。
  mapping_set_unevictable(inode->i_mapping);

  switch (inode->i_mode & S_IFMT) {
  default:
    init_special_inode(inode, inode->i_mode, inode->i_rdev);
    break;
  case S_IFREG:
    inode->i_op = &eltonfs_file_inode_operations;
    inode->i_fop = &eltonfs_file_operations;
    break;
  case S_IFDIR:
    inode->i_op = &eltonfs_dir_inode_operations;
    inode->i_fop = &simple_dir_operations;

    /* directory inodes start off with i_nlink == 2 (for "." entry) */
    inc_nlink(inode);
    break;
  case S_IFLNK:
    inode->i_op = &eltonfs_symlink_inode_operations;
    inode_nohighmem(inode);
    break;
  }
  return eltonfs_i(inode);
}
