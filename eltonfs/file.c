// Private Data of Regular File
//
// file->private_data points a file of local cache.  Should close it if parent
// file closed.
#include <elton/assert.h>
#include <elton/elton.h>
#include <elton/error.h>
#include <elton/local_cache.h>
#include <elton/xattr.h>
#include <linux/cred.h>
#include <linux/mm.h>
#include <linux/mount.h>
#include <linux/namei.h>

static inline struct file *REAL_FILE(struct file *file) {
  return (struct file *)file->private_data;
}
static inline void UPDATE_SIZE_INODE(struct inode *inode) {
  struct inode *real = eltonfs_i(inode)->file.cache_inode;
  i_size_write(inode, i_size_read(real));
  WRITE_ONCE(inode->i_blocks, READ_ONCE(real->i_blocks));
}
static inline void UPDATE_SIZE(struct file *file) {
  UPDATE_SIZE_INODE(file->f_inode);
}
static inline void UPDATE_POS(struct file *from, struct file *to) {
  if (to->f_pos != from->f_pos) {
    to->f_pos = from->f_pos;
    to->f_version = 0;
  }
}

static inline int maybe_load_file(struct inode *inode) {
  int error;
  struct inode *real;
  struct eltonfs_inode *ei = eltonfs_i(inode);
  if (likely(ei->file.cache_inode))
    return 0;
  BUG_ON(ei->file.local_cache_id); // This file created by local.  So cache_id
                                   // should not NULL.
  BUG_ON(!ei->file.object_id);     // cid and oid is NULL.  What is this inode!?

  error = eltonfs_cache_obj(ei->file.object_id, inode->i_sb);
  if (error)
    return error;
  real = eltonfs_get_obj_inode(ei->file.object_id, inode->i_sb);
  if (IS_ERR(real))
    return PTR_ERR(real);
  ei->file.cache_inode = real;
  UPDATE_SIZE_INODE(inode);
  return 0;
}

static int eltonfs_file_mmap(struct file *file, struct vm_area_struct *vma) {
  int ret;
  struct file *real = REAL_FILE(file);
  OBJ_CACHE_ACCESS_START_FILE(file);
  if (!real->f_op->mmap)
    ret = -ENOTSUPP;
  else {
    vma->vm_file = get_file(real);
    ret = real->f_op->mmap(real, vma);
    if (ret)
      // An error occurred.  Callee (mmap_region) will drop the original file
      // during error handling.  This handler should drop the real file.
      fput(real);
    else
      // Finished without error.  Callee (mmap_region) will drop real file. This
      // handler should drop the original file.
      fput(file);
  }
  OBJ_CACHE_ACCESS_END;
  RETURN_EXTERNAL(ret);
}

static int _eltonfs_file_open(struct inode *inode, struct file *file) {
  struct file *real;
  real = eltonfs_open_real_file(eltonfs_i(inode), file);
  if (real && IS_ERR(real))
    return PTR_ERR(real);
  if (real) {
    // Found local cache.
    file->private_data = real;
    UPDATE_SIZE(file);
    UPDATE_POS(real, file);
    return 0;
  }
  return -ELTON_CACHE_MISS;
}
static int eltonfs_file_open(struct inode *inode, struct file *file) {
  int error;
  OBJ_CACHE_ACCESS_START(inode->i_sb);
  if (!(inode->i_mode & S_IFREG)) {
    error = -EINVAL;
    goto out;
  }

  error = _eltonfs_file_open(inode, file);
  if (error != -ELTON_CACHE_MISS)
    // Fast path (cache hit).
    goto out;

  // Slow path (cache miss).
  error = eltonfs_cache_obj(eltonfs_i(inode)->file.object_id, inode->i_sb);
  if (error)
    goto out;
  error = _eltonfs_file_open(inode, file);

out:
  OBJ_CACHE_ACCESS_END;
  RETURN_EXTERNAL(error);
}

static int eltonfs_file_release(struct inode *inode, struct file *file) {
  int error = 0;
  if (file->private_data) {
    OBJ_CACHE_ACCESS_START(inode->i_sb);
    error = filp_close(file->private_data, NULL);
    OBJ_CACHE_ACCESS_END;
  }
  RETURN_EXTERNAL(error);
}

static ssize_t eltonfs_file_read(struct file *file, char __user *buff,
                                 size_t size, loff_t *pos) {
  OBJ_CACHE_ACCESS_START_FILE(file);
  ssize_t ret = vfs_read(REAL_FILE(file), buff, size, pos);
  OBJ_CACHE_ACCESS_END;
  RETURN_EXTERNAL(ret);
}
static ssize_t eltonfs_file_write(struct file *file, const char __user *buff,
                                  size_t size, loff_t *pos) {
  OBJ_CACHE_ACCESS_START_FILE(file);
  ssize_t ret = vfs_write(REAL_FILE(file), buff, size, pos);
  UPDATE_SIZE(file);
  OBJ_CACHE_ACCESS_END;
  RETURN_EXTERNAL(ret);
}
static loff_t eltonfs_file_llseek(struct file *file, loff_t offset,
                                  int whence) {
  OBJ_CACHE_ACCESS_START_FILE(file);
  size_t ret;
  struct file *real = REAL_FILE(file);
  UPDATE_POS(file, real);
  ret = vfs_llseek(real, offset, whence);
  UPDATE_SIZE(file);
  UPDATE_POS(real, file);
  OBJ_CACHE_ACCESS_END;
  RETURN_EXTERNAL(ret);
}
static int eltonfs_file_fsync(struct file *file, loff_t start, loff_t end,
                              int datasync) {
  int ret;
  OBJ_CACHE_ACCESS_START_FILE(file);
  ret = vfs_fsync_range(REAL_FILE(file), start, end, datasync);
  OBJ_CACHE_ACCESS_END;
  RETURN_EXTERNAL(ret);
}
static ssize_t eltonfs_file_splice_read(struct file *in, loff_t *ppos,
                                        struct pipe_inode_info *pipe,
                                        size_t len, unsigned int flags) {
  int ret;
  struct file *real = REAL_FILE(in);
  OBJ_CACHE_ACCESS_START_FILE(in);
  if (!real->f_op->splice_read)
    ret = -ENOTSUPP;
  else
    ret = real->f_op->splice_read(real, ppos, pipe, len, flags);
  OBJ_CACHE_ACCESS_END;
  RETURN_EXTERNAL(ret);
}
static ssize_t eltonfs_file_splice_write(struct pipe_inode_info *pipe,
                                         struct file *out, loff_t *ppos,
                                         size_t len, unsigned int flags) {
  ssize_t ret;
  struct file *real = REAL_FILE(out);
  OBJ_CACHE_ACCESS_START_FILE(out);
  if (!real->f_op->splice_write)
    ret = -ENOTSUPP;
  else {
    ret = real->f_op->splice_write(pipe, real, ppos, len, flags);
    UPDATE_SIZE(out);
  }
  OBJ_CACHE_ACCESS_END;
  RETURN_EXTERNAL(ret);
}
static long eltonfs_file_fallocate(struct file *file, int mode, loff_t offset,
                                   loff_t len) {
  long ret;
  struct file *real = REAL_FILE(file);
  OBJ_CACHE_ACCESS_START_FILE(file);
  if (!real->f_op->fallocate)
    ret = -ENOTSUPP;
  else {
    ret = real->f_op->fallocate(real, mode, offset, len);
    UPDATE_SIZE(file);
  }
  OBJ_CACHE_ACCESS_END;
  RETURN_EXTERNAL(ret);
}

int eltonfs_file_setattr(struct dentry *dentry, struct iattr *iattr) {
  struct inode *inode = d_inode(dentry);
  int error;

  error = setattr_prepare(dentry, iattr);
  if (error)
    return error;

  if (iattr->ia_valid & ATTR_SIZE) {
    struct path path;
    eltonfs_get_cache_path(dentry->d_inode, &path);
    vfs_truncate(&path, iattr->ia_size);
    path_put(&path);
    truncate_setsize(inode, iattr->ia_size);
  }
  setattr_copy(inode, iattr);
  mark_inode_dirty(inode);
  return 0;
}
int eltonfs_file_getattr(const struct path *path, struct kstat *stat,
                         u32 request_mask, unsigned int query_flags) {
  struct inode *inode = d_inode(path->dentry);
  int error = maybe_load_file(inode);
  if (error)
    RETURN_EXTERNAL(error);
  generic_fillattr(inode, stat);
  return 0;
}

const struct file_operations eltonfs_file_operations = {
    .read = eltonfs_file_read,
    .write = eltonfs_file_write,
    .mmap = eltonfs_file_mmap,
    .open = eltonfs_file_open,
    .release = eltonfs_file_release,
    .fsync = eltonfs_file_fsync,
    .splice_read = eltonfs_file_splice_read,
    .splice_write = eltonfs_file_splice_write,
    .llseek = eltonfs_file_llseek,
    .fallocate = eltonfs_file_fallocate,
    .unlocked_ioctl = eltonfs_ioctl,
#ifdef CONFIG_COMPAT
    // for 32bit application.  See
    // https://qiita.com/akachochin/items/94ba679b2941f55c1d2d
    .compat_ioctl = eltonfs_compat_ioctl,
#endif
};

const struct inode_operations eltonfs_file_inode_operations = {
    .setattr = eltonfs_file_setattr,
    .getattr = eltonfs_file_getattr,
#ifdef ELTONFS_XATTRS
    .listxattr = elton_list_xattr_vfs,
#endif
    .update_time = elton_update_time,
};
