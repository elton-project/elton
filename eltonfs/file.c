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
#include <linux/mount.h>

static inline struct file *REAL_FILE(struct file *file) {
  return (struct file *)file->private_data;
}

static int eltonfs_file_mmap(struct file *file, struct vm_area_struct *vma) {
#ifdef ELTONFS_STATISTIC
  struct eltonfs_info *info = file->f_path.mnt->mnt_sb->s_fs_info;
  unsigned long size = vma->vm_end - vma->vm_start;
  int need_logging = 0;

  write_lock(&info->mmap_size_lock);
  if (info->mmap_size < size) {
    info->mmap_size = size;
    need_logging = 1;
  }
  write_unlock(&info->mmap_size_lock);

  if (need_logging)
    DEBUG("mmap size: file=%s, size=%ld", file->f_path.dentry->d_name.name,
          size);
#endif

  return generic_file_mmap(file, vma);
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
  return current->mm->get_unmapped_area(file, addr, len, pgoff, flags);
}

ssize_t eltonfs_file_read(struct file *file, char __user *buff, size_t size,
                          loff_t *offset) {
  OBJ_CACHE_ACCESS_START_FILE(file);
  ssize_t ret = vfs_read(REAL_FILE(file), buff, size, offset);
  OBJ_CACHE_ACCESS_END;
  return ret;
}

struct file_operations eltonfs_file_operations = {
    .read = eltonfs_file_read,
    .write_iter = generic_file_write_iter,
    .mmap = eltonfs_file_mmap,
    .open = eltonfs_file_open,
    .release = eltonfs_file_release,
    .fsync = noop_fsync,
    .splice_read = generic_file_splice_read,
    .splice_write = iter_file_splice_write,
    .llseek = generic_file_llseek,
    .get_unmapped_area = eltonfs_get_unmapped_area,
    .unlocked_ioctl = eltonfs_ioctl,
#ifdef CONFIG_COMPAT
    // for 32bit application.  See
    // https://qiita.com/akachochin/items/94ba679b2941f55c1d2d
    .compat_ioctl = eltonfs_compat_ioctl,
#endif
};

struct inode_operations eltonfs_file_inode_operations = {
    .setattr = simple_setattr,
    .getattr = simple_getattr,
#ifdef ELTONFS_XATTRS
    .listxattr = elton_list_xattr_vfs,
#endif
    .update_time = elton_update_time,
};
