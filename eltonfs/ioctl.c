#include <elton/commit.h>
#include <elton/elton.h>
#include <elton/rpc/server.h>
#include <elton/rpc/struct.h>
#include <linux/uaccess.h>

long eltonfs_ioctl_commit(struct super_block *sb) {
  struct inode *root = sb->s_root->d_inode;
  struct tree_info *new_tree = NULL;
  const char *new_cid = NULL;
  const char *old_cid = NULL;
  struct commit_info *old_cinfo = NULL;
  struct commit_info *new_cinfo = NULL;

  // todo: block write access.
  new_tree = eltonfs_build_tree(root);
  if (IS_ERR(new_tree)) {
    return PTR_ERR(new_tree);
  }

  new_cid = eltonfs_call_commit(sb, new_tree);
  if (IS_ERR(new_cid)) {
    // todo: free new_tree
    return PTR_ERR(new_cid);
  }
  new_cinfo = eltonfs_get_commit(new_cid);
  if (IS_ERR(new_cid)) {
    // todo: free new tree
    return PTR_ERR(new_cinfo);
  }

  // todo: block all access.
  // Change to latest commit.
  old_cid = eltonfs_sb(sb)->cid;
  old_cinfo = eltonfs_sb(sb)->cinfo;
  eltonfs_sb(sb)->cid = new_cid;
  eltonfs_sb(sb)->cinfo = new_cinfo;
  eltonfs_sb(sb)->inodes_ei = new_cinfo->tree->inodes;

  eltonfs_apply_tree(root, new_cinfo->tree);
  // todo: unblock all access.

  kfree(old_cid);
  // todo: free old_cinfo
  return 0;
}

long eltonfs_ioctl(struct file *file, unsigned int cmd, unsigned long arg) {
  struct inode *inode = file_inode(file);
  struct super_block *sb = inode->i_sb;

  switch (cmd) {
  case ELTONFS_IOC_BEGIN:
  case ELTONFS_IOC_ROLLBACK:
    // todo
    break;
  case ELTONFS_IOC_COMMIT:
    return eltonfs_ioctl_commit(sb);
  case FS_IOC_GETFLAGS: {
    unsigned int flags = 0;
    // TODO: 拡張属性に対応する。
    return put_user(flags, (int __user *)arg);
  }
  case FS_IOC_GETVERSION:
    return put_user(inode->i_generation, (int __user *)arg);
  }
  return -ENOTTY; // Not implemented
}

#ifdef CONFIG_COMPAT
long eltonfs_compat_ioctl(struct file *file, unsigned int cmd,
                          unsigned long arg) {
  switch (cmd) {
  case FS_IOC32_GETFLAGS:
    cmd = FS_IOC_GETFLAGS;
    break;
  case FS_IOC32_GETVERSION:
    cmd = FS_IOC_GETVERSION;
    break;
  }
  return eltonfs_ioctl(file, cmd, arg);
}
#endif
