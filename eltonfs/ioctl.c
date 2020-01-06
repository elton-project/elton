#include <elton/elton.h>
#include <linux/uaccess.h>

long eltonfs_ioctl_commit(struct super_block *sb) {
  // todo: block write access.
  // todo: build new tree from vfs_inode.
  // todo: send tree.
  // todo: wait response.
  // todo: block all access.
  // todo: change commit.
  // todo: unblock all access.
}

long eltonfs_ioctl(struct file *file, unsigned int cmd, unsigned long arg) {
  struct inode *inode = file_inode(file);
  struct super_block *sb = inode->i_sb;

  switch (cmd) {
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
