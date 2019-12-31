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

static struct dentry *eltonfs_lookup(struct inode *vfs_dir,
                                     struct dentry *dentry,
                                     unsigned int flags) {
  const struct eltonfs_inode *dir = eltonfs_i(vfs_dir);
  const char *name = dentry->d_name.name;
  const size_t name_len = strlen(name);
  struct eltonfs_dir_entry *entry;
  struct inode *inode;

  if (name_len > ELTONFS_NAME_LEN)
    return ERR_PTR(-ENAMETOOLONG);

  list_for_each_entry(entry, &dir->dir.dir_entries._list_head, _list_head) {
    if (entry->name_len != name_len)
      // Fast path
      continue;
    // Slow path
    if (strncmp(entry->name, name, entry->name_len))
      continue;

    // Found
    inode = vfs_i(eltonfs_iget(vfs_dir->i_sb, entry->ino));
    if (IS_ERR(inode))
      return ERR_CAST(inode);
    if (!inode) {
      WARN_ONCE(1, "a directory entry referring to a non-existent inode is "
                   "found.  Internal tree is corrupted!");
      // todo: change error code.
      return ERR_PTR(-EINVAL);
    }
    return d_splice_alias(inode, dentry);
  }
  // Not found
  return ERR_PTR(-ENOENT);
}

// todo
struct file_operations eltonfs_dir_operations = {
    .llseek = NULL,
    .read = NULL,
    .write = NULL,
    .iterate_shared = NULL,
    .unlocked_ioctl = NULL,
#ifdef CONFIG_COMPAT
    .compat_ioctl = NULL,
#endif
    .fsync = NULL,
};

struct inode_operations eltonfs_dir_inode_operations = {
    .create = eltonfs_create,
    .lookup = eltonfs_lookup,
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
