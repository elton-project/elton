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

static int eltonfs_file_mmap(struct file *file, struct vm_area_struct *vma) {
  struct file *real = REAL_FILE(file);
  if (!real->f_op->mmap)
    return -ENOTSUPP;
  vma->vm_file = real;
  return real->f_op->mmap(real, vma);
}

static int _eltonfs_file_open(struct inode *inode, struct file *file) {
  struct file *real;
  real = eltonfs_open_real_file(eltonfs_i(inode), file);
  if (real && IS_ERR(real))
    return PTR_ERR(real);
  if (real) {
    // Found local cache.
    file->private_data = real;
    return 0;
  }
  return -ELTON_CACHE_MISS;
}
static int eltonfs_file_open(struct inode *inode, struct file *file) {
  int error;

  if (!(inode->i_mode & S_IFREG))
    return -EINVAL;

  error = _eltonfs_file_open(inode, file);
  if (error != -ELTON_CACHE_MISS)
    // Fast path (cache hit).
    return error;

  // Slow path (cache miss).
  error = eltonfs_cache_obj(eltonfs_i(inode)->file.object_id, inode->i_sb);
  if (error)
    return error;
  return _eltonfs_file_open(inode, file);
}

static int eltonfs_file_release(struct inode *inode, struct file *file) {
  if (file->private_data)
    return filp_close(file->private_data, NULL);
  return 0;
}

static unsigned long eltonfs_get_unmapped_area(struct file *file,
                                               unsigned long addr,
                                               unsigned long len,
                                               unsigned long pgoff,
                                               unsigned long flags) {
  struct file *real = REAL_FILE(file);
  if (!real->f_op->get_unmapped_area)
    return -ENOTSUPP;
  return real->f_op->get_unmapped_area(real, addr, len, pgoff, flags);
}

static ssize_t eltonfs_file_read(struct file *file, char __user *buff,
                                 size_t size, loff_t *offset) {
  OBJ_CACHE_ACCESS_START_FILE(file);
  ssize_t ret = vfs_read(REAL_FILE(file), buff, size, offset);
  OBJ_CACHE_ACCESS_END;
  return ret;
}
static ssize_t eltonfs_file_write(struct file *file, const char __user *buff,
                                  size_t size, loff_t *offset) {
  OBJ_CACHE_ACCESS_START_FILE(file);
  size_t ret = vfs_write(REAL_FILE(file), buff, size, offset);
  OBJ_CACHE_ACCESS_END;
  // todo: update size.
  return ret;
}
static loff_t eltonfs_file_llseek(struct file *file, loff_t offset,
                                  int whence) {
  return vfs_llseek(REAL_FILE(file), offset, whence);
}
static int eltonfs_file_fsync(struct file *file, loff_t start, loff_t end,
                              int datasync) {
  return vfs_fsync_range(REAL_FILE(file), start, end, datasync);
}
static ssize_t eltonfs_file_splice_read(struct file *in, loff_t *ppos,
                                        struct pipe_inode_info *pipe,
                                        size_t len, unsigned int flags) {
  struct file *real = REAL_FILE(in);
  if (!real->f_op->splice_read)
    return -ENOTSUPP;
  return real->f_op->splice_read(real, ppos, pipe, len, flags);
}
static ssize_t eltonfs_file_splice_write(struct pipe_inode_info *pipe,
                                         struct file *out, loff_t *ppos,
                                         size_t len, unsigned int flags) {
  struct file *real = REAL_FILE(out);
  if (!real->f_op->splice_write)
    return -ENOTSUPP;
  return real->f_op->splice_write(pipe, real, ppos, len, flags);
}
static long eltonfs_file_fallocate(struct file *file, int mode, loff_t offset,
                                   loff_t len) {
  struct file *real = REAL_FILE(file);
  if (!real->f_op->fallocate)
    return -ENOTSUPP;
  return real->f_op->fallocate(real, mode, offset, len);
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
  generic_fillattr(inode, stat);
  return 0;
}

struct file_operations eltonfs_file_operations = {
    .read = eltonfs_file_read,
    .write = eltonfs_file_write,
    .mmap = eltonfs_file_mmap,
    .open = eltonfs_file_open,
    .release = eltonfs_file_release,
    .fsync = eltonfs_file_fsync,
    .splice_read = eltonfs_file_splice_read,
    .splice_write = eltonfs_file_splice_write,
    .llseek = eltonfs_file_llseek,
    .get_unmapped_area = eltonfs_get_unmapped_area,
    .fallocate = eltonfs_file_fallocate,
    .unlocked_ioctl = eltonfs_ioctl,
#ifdef CONFIG_COMPAT
    // for 32bit application.  See
    // https://qiita.com/akachochin/items/94ba679b2941f55c1d2d
    .compat_ioctl = eltonfs_compat_ioctl,
#endif
};

struct inode_operations eltonfs_file_inode_operations = {
    .setattr = eltonfs_file_setattr,
    .getattr = eltonfs_file_getattr,
#ifdef ELTONFS_XATTRS
    .listxattr = elton_list_xattr_vfs,
#endif
    .update_time = elton_update_time,
};
