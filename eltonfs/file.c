#include <elton/assert.h>
#include <elton/elton.h>
#include <elton/xattr.h>
#include <linux/mount.h>

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

static unsigned long eltonfs_get_unmapped_area(struct file *file,
                                               unsigned long addr,
                                               unsigned long len,
                                               unsigned long pgoff,
                                               unsigned long flags) {
  return current->mm->get_unmapped_area(file, addr, len, pgoff, flags);
}

struct file_operations eltonfs_file_operations = {
    .read_iter = generic_file_read_iter,
    .write_iter = generic_file_write_iter,
    .mmap = eltonfs_file_mmap,
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
