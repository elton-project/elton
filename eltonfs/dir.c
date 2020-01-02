// Private Data of Directory File
//
// To iterate directory entries, file->private_data represents an index value
// (unsigned long) instead of memory address.
//
// Opened:
//     file->private_data should (void *)0.
// iterate_shared() is called:
//     file->private_data represents index of next directory entry.
// Closed:
//     Not defined.
#include <elton/assert.h>
#include <elton/elton.h>
#include <elton/xattr.h>

static inline struct file *_eltonfs_real_file(struct file *file,
                                              const char *caller) {
  if (!file->private_data) {
    DEBUG("%s: private_data is null: file=%px", caller, file);
    BUG();
  }
  return file->private_data;
}
#define REAL_FILE(file) _eltonfs_real_file((file), __func__)

static inline struct eltonfs_dir_entry *
_eltonfs_dir_entries_lookup_entry(struct inode *vfs_dir, const char *name) {
  const struct eltonfs_inode *dir = eltonfs_i(vfs_dir);
  const size_t name_len = strlen(name);
  struct eltonfs_dir_entry *entry;

  list_for_each_entry(entry, &dir->dir.dir_entries._list_head, _list_head) {
    if (likely(entry->name_len != name_len))
      // Fast path
      continue;
    // Slow path
    if (likely(strncmp(entry->name, name, entry->name_len)))
      continue;
    return entry;
  }
  // Not found
  return NULL;
}

// Lookup an entry by name from dir.
// It returns vfs_ino or 0.
static u64 eltonfs_dir_entries_lookup(struct inode *dir, const char *name) {
  struct eltonfs_dir_entry *entry =
      _eltonfs_dir_entries_lookup_entry(dir, name);
  if (unlikely(!entry))
    return 0;
  return entry->ino;
}
// Delete an entry from dir.
static int eltonfs_dir_entries_delete(struct inode *dir, const char *name) {
  struct eltonfs_dir_entry *entry =
      _eltonfs_dir_entries_lookup_entry(dir, name);
  if (unlikely(!entry))
    return -ENOENT;

  list_del(&entry->_list_head);
  kfree(entry);
  eltonfs_i(dir)->dir.count--;
}
// Add an entry with specified name and vfs_ino.
static int eltonfs_dir_entries_add(struct inode *dir, const char *name,
                                   u64 vfs_ino) {
  struct eltonfs_dir_entry *entry;
  size_t len;
  len = strlen(name);
  if (unlikely(len > ELTONFS_NAME_LEN))
    return -ENAMETOOLONG;

  entry = kmalloc(sizeof(*entry), GFP_NOFS);
  if (unlikely(!entry))
    return -ENOMEM;
  entry->ino = vfs_ino;
  memcpy(entry->name, name, len);
  entry->name_len = len;
  list_add(&entry->_list_head, &eltonfs_i(dir)->dir.dir_entries._list_head);
  return 0;
}

static int eltonfs_dir_open(struct inode *inode, struct file *file) {
  file->private_data = (void *)0;
  return 0;
}

static int eltonfs_iterate_shared(struct file *file, struct dir_context *ctx) {
  struct eltonfs_inode *ei = eltonfs_i(file->f_inode);
  struct eltonfs_dir_entry *entry;
  unsigned long seek_index;
  unsigned long index; // Index of next directory entry.
  BUILD_BUG_ON(sizeof(void *) != sizeof(unsigned long));

  index = (long)file->private_data;
  if (ei->dir.count < index)
    // Reached to end of directory entries list.
    return 0;

  if (index == 0) {
    index = 1;
    if (!dir_emit_dot(file, ctx))
      goto out;
  }
  if (index == 1) {
    index = 2;
    if (!dir_emit_dot(file, ctx))
      goto out;
  }

  seek_index = 2;
  ELTONFS_FOR_EACH_DIRENT(ei, entry) {
    if (seek_index < index) {
      seek_index++;
      continue;
    }
    seek_index++;
    index++;
    if (!dir_emit(ctx, entry->name, entry->name_len, entry->ino, DT_UNKNOWN))
      goto out;
  }

out:
  file->private_data = (void *)index;
  return 0;
}

static int eltonfs_dir_release(struct inode *inode, struct file *file) {
  return 0;
}

long eltonfs_unlocked_ioctl(struct file *, unsigned int, unsigned long);
long eltonfs_compat_ioctl(struct file *, unsigned int, unsigned long);

static int eltonfs_dir_fsync(struct file *file, loff_t start, loff_t end,
                             int datasync) {
  // Directory is not associate to real file.
  return -ENOTSUPP;
}

static int eltonfs_mknod(struct inode *dir, struct dentry *dentry, umode_t mode,
                         dev_t dev) {
  struct inode *inode = eltonfs_create_inode(dir->i_sb, dir, mode, dev);
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

  inode = eltonfs_create_inode(dir->i_sb, dir, S_IFLNK | S_IRWXUGO, 0);
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
    atomic_inc(&inode->i_count);
    return d_splice_alias(inode, dentry);
  }
  // Not found
  d_add(dentry, NULL);
  return dentry;
}

// todo
struct file_operations eltonfs_dir_operations = {
    .open = eltonfs_dir_open,
    .release = eltonfs_dir_release,
    .iterate_shared = eltonfs_iterate_shared,
// todo
// .unlocked_ioctl = eltonfs_unlocked_ioctl,
#ifdef CONFIG_COMPAT
// todo
// .compat_ioctl = eltonfs_compat_ioctl,
#endif
    .fsync = eltonfs_dir_fsync,
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
