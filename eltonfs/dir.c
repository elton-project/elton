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
#include <elton/utils.h>
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

  list_for_each_entry(entry, &dir->dir.dir_entries, _list_head) {
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
  struct eltonfs_dir_entry *entry;
  u64 ino;

  spin_lock(&eltonfs_i(dir)->lock);
  entry = _eltonfs_dir_entries_lookup_entry(dir, name);
  if (unlikely(!entry))
    ino = 0;
  else
    ino = entry->ino;
  spin_unlock(&eltonfs_i(dir)->lock);
  return ino;
}
// Delete an entry from dir.
static int eltonfs_dir_entries_delete(struct inode *dir, const char *name) {
  struct eltonfs_dir_entry *entry;
  int error = 0;

  spin_lock(&eltonfs_i(dir)->lock);
  entry = _eltonfs_dir_entries_lookup_entry(dir, name);
  if (unlikely(!entry)) {
    error = -ENOENT;
    goto out;
  }

  list_del(&entry->_list_head);
  kfree(entry);
  eltonfs_i(dir)->dir.count--;
  dir->i_mtime = dir->i_ctime = current_time(dir);

out:
  spin_unlock(&eltonfs_i(dir)->lock);
  return error;
}
// Add an entry with specified name and vfs_ino.
static int eltonfs_dir_entries_add(struct inode *dir, const char *name,
                                   u64 vfs_ino) {
  int error = 0;
  struct eltonfs_dir_entry *entry;
  size_t len;
  len = strlen(name);
  if (unlikely(len > ELTONFS_NAME_LEN))
    return -ENAMETOOLONG;

  spin_lock(&eltonfs_i(dir)->lock);

  entry = kmalloc(sizeof(*entry), GFP_NOFS);
  if (unlikely(!entry)) {
    error = -ENOMEM;
    goto out;
  }
  entry->ino = vfs_ino;
  memcpy(entry->name, name, len);
  entry->name_len = len;
  list_add(&entry->_list_head, &eltonfs_i(dir)->dir.dir_entries);
  eltonfs_i(dir)->dir.count++;
  dir->i_mtime = dir->i_ctime = current_time(dir);

out:
  spin_unlock(&eltonfs_i(dir)->lock);
  return 0;
}
// Replace a directory entry.
// The destination file does not have to exist.
static int eltonfs_dir_entries_replace(struct inode *old_dir,
                                       const char *old_name,
                                       struct inode *new_dir,
                                       const char *new_name) {
  struct eltonfs_dir_entry *old_entry, *new_entry;
  struct eltonfs_inode *i_low, *i_high;

  // Lock acquiring order is decided by address.
  if (old_dir < new_dir) {
    i_low = eltonfs_i(old_dir);
    i_high = eltonfs_i(new_dir);
  } else if (old_dir == new_dir) {
    // old_dir and new_dir is a same directory.
    i_low = eltonfs_i(old_dir);
    i_high = NULL;
  } else {
    i_low = eltonfs_i(new_dir);
    i_high = eltonfs_i(old_dir);
  }
  spin_lock(&i_low->lock);
  if (i_high)
    spin_lock(&i_high->lock);

  old_entry = _eltonfs_dir_entries_lookup_entry(old_dir, old_name);
  new_entry = _eltonfs_dir_entries_lookup_entry(new_dir, new_name);

  if (new_entry) {
    list_del(&new_entry->_list_head); // Disconnect from new_dir.
    eltonfs_i(new_dir)->dir.count--;
    kfree(new_entry);
  }
  list_del(&old_entry->_list_head); // Disconnect from old_dir.
  eltonfs_i(old_dir)->dir.count--;
  strcpy(old_entry->name, new_name);
  old_entry->name_len = strlen(new_name);
  list_add(&old_entry->_list_head, &eltonfs_i(new_dir)->dir.dir_entries);
  eltonfs_i(new_dir)->dir.count++;

  if (i_high)
    spin_unlock(&i_high->lock);
  spin_unlock(&i_low->lock);
  return 0;
}
static int eltonfs_dir_entries_is_empty(struct inode *dir) {
  int ret;
  spin_lock(&eltonfs_i(dir)->lock);
  ret = list_empty(&eltonfs_i(dir)->dir.dir_entries);
  spin_unlock(&eltonfs_i(dir)->lock);
  return ret;
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

  spin_lock(&ei->lock);
  index = (long)file->private_data;
  if (ei->dir.count < index)
    // Reached to end of directory entries list.
    goto out_without_updating_index;

  if (index == 0) {
    index = 1;
    if (!dir_emit_dot(file, ctx))
      goto out;
  }
  if (index == 1) {
    index = 2;
    if (!dir_emit_dotdot(file, ctx))
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
out_without_updating_index:
  spin_unlock(&ei->lock);
  return 0;
}

static int eltonfs_dir_release(struct inode *inode, struct file *file) {
  return 0;
}

static int eltonfs_dir_fsync(struct file *file, loff_t start, loff_t end,
                             int datasync) {
  // Directory is not associate to real file.
  return -ENOTSUPP;
}

// Create regular/directory/device file into dir.
static int eltonfs_mknod(struct inode *dir, struct dentry *dentry, umode_t mode,
                         dev_t dev) {
  int error;
  struct inode *inode = eltonfs_create_inode(dir->i_sb, dir, mode, dev);
  if (!inode)
    return -ENOSPC;

  switch (mode & S_IFMT) {
  case S_IFREG:
    eltonfs_inode_init_regular_with_new_cache(inode);
    break;
  case S_IFDIR:
    eltonfs_inode_init_dir(inode);
    break;
  }

  error = eltonfs_dir_entries_add(dir, dentry->d_name.name, inode->i_ino);
  if (error)
    RETURN_EXTERNAL(error);

  d_instantiate(dentry, inode);
  dget(dentry);
  return 0;
}

static int eltonfs_create(struct inode *dir, struct dentry *dentry,
                          umode_t mode, bool excl) {
  RETURN_EXTERNAL(eltonfs_mknod(dir, dentry, mode | S_IFREG, 0));
}

static int eltonfs_symlink(struct inode *dir, struct dentry *dentry,
                           const char *symname) {
  int error;
  struct inode *inode;

  inode = eltonfs_create_inode(dir->i_sb, dir, S_IFLNK | S_IRWXUGO, 0);
  if (!inode)
    return -ENOSPC;

  eltonfs_inode_init_symlink(inode, NULL, symname);

  error = eltonfs_dir_entries_add(dir, dentry->d_name.name, inode->i_ino);
  if (error)
    RETURN_EXTERNAL(error);

  d_instantiate(dentry, inode);
  dget(dentry);
  dir->i_mtime = dir->i_ctime = current_time(dir);
  return 0;
}

static int eltonfs_mkdir(struct inode *dir, struct dentry *dentry,
                         umode_t mode) {
  int error = eltonfs_mknod(dir, dentry, mode | S_IFDIR, 0);
  if (error)
    RETURN_EXTERNAL(error);
  inc_nlink(dir);
  return 0;
}

static struct dentry *eltonfs_lookup(struct inode *vfs_dir,
                                     struct dentry *dentry,
                                     unsigned int flags) {
  u64 ino = eltonfs_dir_entries_lookup(vfs_dir, dentry->d_name.name);
  if (likely(ino)) {
    // Found
    struct inode *inode = vfs_i(eltonfs_iget(vfs_dir->i_sb, ino));
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
  return NULL;
}

int eltonfs_link(struct dentry *old_dentry, struct inode *new_dir,
                 struct dentry *new_dentry) {
  struct inode *inode = d_inode(old_dentry);

  eltonfs_dir_entries_add(new_dir, new_dentry->d_name.name, inode->i_ino);

  inode->i_ctime = new_dir->i_ctime = new_dir->i_mtime = current_time(inode);
  inc_nlink(inode);
  ihold(inode);
  dget(new_dentry);
  d_instantiate(new_dentry, inode);
  return 0;
}

int eltonfs_unlink(struct inode *dir, struct dentry *dentry) {
  int error;
  struct inode *inode = d_inode(dentry);

  error = eltonfs_dir_entries_delete(dir, dentry->d_name.name);
  if (error)
    RETURN_EXTERNAL(error);
  drop_nlink(inode);
  dput(dentry);
  return 0;
}

int eltonfs_rmdir(struct inode *dir, struct dentry *dentry) {
  struct inode *child = d_inode(dentry);
  if ((child->i_mode & S_IFDIR) && !eltonfs_dir_entries_is_empty(child))
    // Directory is not empty.
    return -ENOTEMPTY;

  drop_nlink(child);
  eltonfs_unlink(dir, dentry);
  drop_nlink(dir);
  return 0;
}

int eltonfs_rename(struct inode *old_dir, struct dentry *old_dentry,
                   struct inode *new_dir, struct dentry *new_dentry,
                   unsigned int flags) {
  struct inode *old = d_inode(old_dentry);
  struct inode *new = d_inode(new_dentry);

  if (flags & ~RENAME_NOREPLACE)
    return -EINVAL;
  // NOTE: The VFS already checks for existence, so for eltonfs the
  // RENAME_NOREPLACE implementation is not needed.

  if (d_really_is_positive(new_dentry)) {
    // Already exists a file to destination.  Should remove dest file before
    // move a file.
    if (d_is_dir(new_dentry)) {
      if (!eltonfs_dir_entries_is_empty(new))
        return -ENOTEMPTY;
    }

    eltonfs_dir_entries_replace(old_dir, old_dentry->d_name.name, new_dir,
                                new_dentry->d_name.name);
    drop_nlink(old_dir);
    drop_nlink(new);
  } else {
    // Move a file.
    eltonfs_dir_entries_replace(old_dir, old_dentry->d_name.name, new_dir,
                                new_dentry->d_name.name);
    drop_nlink(old_dir);
    inc_nlink(new_dir);
  }

  old_dir->i_ctime = old_dir->i_mtime = new_dir->i_ctime = new_dir->i_mtime =
      old->i_ctime = current_time(old_dir);
  return 0;
}

// todo
const struct file_operations eltonfs_dir_operations = {
    .open = eltonfs_dir_open,
    .release = eltonfs_dir_release,
    .iterate_shared = eltonfs_iterate_shared,
    .unlocked_ioctl = eltonfs_ioctl,
#ifdef CONFIG_COMPAT
    .compat_ioctl = eltonfs_compat_ioctl,
#endif
    .fsync = eltonfs_dir_fsync,
};

const struct inode_operations eltonfs_dir_inode_operations = {
    .create = eltonfs_create,
    .lookup = eltonfs_lookup,
    .link = eltonfs_link,
    .unlink = eltonfs_unlink,
    .symlink = eltonfs_symlink,
    .mkdir = eltonfs_mkdir,
    .rmdir = eltonfs_rmdir,
    .mknod = eltonfs_mknod,
    .rename = eltonfs_rename,
#ifdef ELTONFS_XATTRS
    .listxattr = elton_list_xattr_vfs,
#endif
    .update_time = elton_update_time,
};
