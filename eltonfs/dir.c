#include <elton/elton.h>
#include <elton/xattr.h>

static int eltonfs_mknod(struct inode *dir, struct dentry *dentry, umode_t mode,
                         dev_t dev) {
  struct inode *inode = eltonfs_get_inode(dir->i_sb, dir, mode, dev);
  if (!inode) {
    return -ENOSPC;
  }
  d_instantiate(dentry, inode);
  dget(dentry);
  dir->i_mtime = dir->i_ctime = current_time(dir);
  return 0;
}

static int eltonfs_create(struct inode *dir, struct dentry *dentry,
                          umode_t mode, bool excl) {
  return eltonfs_mknod(dir, dentry, mode | S_IFREG, 0);
}

static int eltonfs_symlink(struct inode *dir, struct dentry *dentry,
                           const char *symname) {
  struct inode *inode;
  int len, error;

  inode = eltonfs_get_inode(dir->i_sb, dir, S_IFLNK | S_IRWXUGO, 0);
  if (!inode) {
    return -ENOSPC;
  }
  len = strlen(symname) + 1;
  // TODO: allocate physical pages.
  error = page_symlink(inode, symname, len);
  if (error) {
    iput(inode);
  }
  d_instantiate(dentry, inode);
  dget(dentry);
  dir->i_mtime = dir->i_ctime = current_time(dir);
  return 0;
}

static int eltonfs_mkdir(struct inode *dir, struct dentry *dentry,
                         umode_t mode) {
  int error = eltonfs_mknod(dir, dentry, mode | S_IFDIR, 0);
  if (error) {
    return error;
  }
  inc_nlink(dir);
  return 0;
}

struct inode_operations eltonfs_dir_inode_operations = {
    .create = eltonfs_create,
    // on-memory
    // filesystemでは、保持しているファイルに対応するdentryは、必ず存在する。
    // オンメモリファイルシステムでのlookupが呼び出されるタイミングは、存在しないファイル
    // にアクセスしたときだけである。
    // このため、simple_lookupは常にnegative dentryを返す。
    //
    // on-disk filesystemでは、lookup関数を自前実装する必要がある。
    .lookup = simple_lookup,
    .link = simple_link,
    .unlink = simple_unlink,
    .symlink = eltonfs_symlink,
    .mkdir = eltonfs_mkdir,
    .rmdir = simple_rmdir,
    .mknod = eltonfs_mknod,
    .rename = simple_rename,
#ifdef ELTONFS_XATTRS
    .listxattr = elton_list_xattr_vfs,
#endif
    .update_time = elton_update_time,
};
