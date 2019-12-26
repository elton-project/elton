#include <elton/assert.h>
#include <elton/elton.h>
#include <elton/rpc/server.h>
#include <elton/rpc/test.h>
#include <elton/xattr.h>
#include <elton/xdr/test.h>
#include <linux/dcache.h>
#include <linux/fs.h>
#include <linux/kernel.h>
#include <linux/module.h>
#include <linux/mount.h>
#include <linux/net.h>
#include <linux/pagemap.h>
#include <linux/seq_file.h>
#include <linux/slab.h>
#include <linux/statfs.h>

static bool is_registered = 0;
static struct elton_rpc_server server;
static struct file_system_type eltonfs_type;
static struct super_operations eltonfs_s_op;
static struct address_space_operations eltonfs_aops;
static struct inode_operations eltonfs_file_inode_operations;
static struct inode_operations eltonfs_dir_inode_operations;
static struct inode_operations eltonfs_symlink_inode_operations;
static struct file_operations eltonfs_file_operations;

static struct inode *eltonfs_get_inode(struct super_block *sb,
                                       const struct inode *dir, umode_t mode,
                                       dev_t dev) {
  struct inode *inode;
  inode = new_inode(sb);
  if (!inode) {
    return inode;
  }

  inode->i_ino = get_next_ino();
  inode_init_owner(inode, dir, mode);
  inode->i_mapping->a_ops = &eltonfs_aops;
  mapping_set_gfp_mask(inode->i_mapping, GFP_HIGHUSER);
  mapping_set_unevictable(inode->i_mapping);
  inode->i_atime = inode->i_mtime = inode->i_ctime = current_time(inode);
  switch (mode & S_IFMT) {
  default:
    init_special_inode(inode, mode, dev);
    break;
  case S_IFREG:
    inode->i_op = &eltonfs_file_inode_operations;
    inode->i_fop = &eltonfs_file_operations;
    break;
  case S_IFDIR:
    inode->i_op = &eltonfs_dir_inode_operations;
    inode->i_fop = &simple_dir_operations;

    /* directory inodes start off with i_nlink == 2 (for "." entry) */
    inc_nlink(inode);
    break;
  case S_IFLNK:
    inode->i_op = &eltonfs_symlink_inode_operations;
    inode_nohighmem(inode);
    break;
  }
  return inode;
}

static int elton_update_time(struct inode *inode, struct timespec64 *time,
                             int flags) {
  spin_lock(&inode->i_lock);
  if (flags & S_ATIME)
    inode->i_atime = *time;
  // Elton is not supported the inode->i_version.
  // if (flags & S_VERSION)
  // 	inode_maybe_inc_iversion(inode, false);
  if (flags & S_CTIME)
    inode->i_ctime = *time;
  if (flags & S_MTIME)
    inode->i_mtime = *time;
  spin_unlock(&inode->i_lock);
  return 0;
}

static int eltonfs_set_page_dirty(struct page *page) {
  if (PageDirty(page)) {
    return 0;
  }
  SetPageDirty(page);
  return 0;
}

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

static int eltonfs_mkdir(struct inode *dir, struct dentry *dentry,
                         umode_t mode) {
  int error = eltonfs_mknod(dir, dentry, mode | S_IFDIR, 0);
  if (error) {
    return error;
  }
  inc_nlink(dir);
  return 0;
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

static unsigned long eltonfs_get_unmapped_area(struct file *file,
                                               unsigned long addr,
                                               unsigned long len,
                                               unsigned long pgoff,
                                               unsigned long flags) {
  return current->mm->get_unmapped_area(file, addr, len, pgoff, flags);
}

static long eltonfs_ioctl(struct file *file, unsigned int cmd,
                          unsigned long arg) {
  struct inode *inode = file_inode(file);
  unsigned int flags;

  switch (cmd) {
  case FS_IOC_GETFLAGS: {
    // TODO: 拡張属性に対応する。
    flags = 0;
    return put_user(flags, (int __user *)arg);
  }
  case FS_IOC_GETVERSION:
    return put_user(inode->i_generation, (int __user *)arg);
  }
  return -ENOSYS; // Not implemented
}

#ifdef CONFIG_COMPAT
static long eltonfs_compat_ioctl(struct file *file, unsigned int cmd,
                                 unsigned long arg) {
  switch (cmd) {
  case FS_IOC32_GETFLAGS:
    cmd = FS_IOC_GETFLAGS;
    break;
  case FS_IOC32_GETVERSION:
    cmd = FS_IOC_GETVERSION;
    break;
  default:
    return -ENOSYS;
  }
  return eltonfs_ioctl(file, cmd, arg);
}
#endif

static int eltonfs_fill_super(struct super_block *sb, void *data, int silent) {
  struct inode *inode;
  struct dentry *root;
  struct iattr ia;

  struct eltonfs_info *info = kmalloc(sizeof(struct eltonfs_info), GFP_KERNEL);
#ifdef ELTONFS_STATISTIC
  rwlock_init(&info->mmap_size_lock);
  info->mmap_size = 0;
#endif

  DEBUG("Preparing for super block ...");
  sb->s_blocksize_bits = PAGE_SHIFT;
  sb->s_blocksize = PAGE_SIZE;
  sb->s_maxbytes = MAX_LFS_FILESIZE;
  sb->s_type = &eltonfs_type;
  sb->s_op = &eltonfs_s_op;
  sb->s_time_gran = 1;
  sb->s_fs_info = info;
#ifdef ELTONFS_XATTRS
  sb->s_xattr = elton_xattr_handlers;
#endif

  inode = eltonfs_get_inode(sb, NULL, S_IFDIR, 0);
  ASSERT_NOT_NULL(inode);
  // Set directory mode to 755;
  inode_lock(inode);
  ia.ia_valid = ATTR_MODE;
  ia.ia_mode = (inode->i_mode & S_IFMT) | 0755;
  setattr_copy(inode, &ia);
  inode_unlock(inode);

  root = d_make_root(inode);
  ASSERT_NOT_NULL(root);
  sb->s_root = root;
  DEBUG("Prepared the super block");
  return 0;
}
static struct dentry *mount(struct file_system_type *fs_type, int flags,
                            const char *dev_name, void *data) {
  return mount_nodev(fs_type, flags, data, eltonfs_fill_super);
}
static void kill_sb(struct super_block *sb) {}

struct inode *eltonfs_alloc_inode(struct super_block *sb) {
  struct eltonfs_inode *i = kmalloc(sizeof(struct eltonfs_inode), GFP_KERNEL);
  if (i == NULL)
    return NULL;

  eltonfs_inode_init_once(i);
  return vfs_i(i);
}
void eltonfs_inode_init_once(struct eltonfs_inode *i) {
#ifdef ELTONFS_XATTRS
  simple_xattrs_init(&i->xattrs);
#endif

  inode_init_once(vfs_i(i));
}
static void eltonfs_destory_inode(struct inode *inode) {
  struct eltonfs_inode *i = eltonfs_i(inode);

#ifdef ELTONFS_XATTRS
  simple_xattrs_free(&i->xattrs);
#endif

  kfree(i);
}

static int eltonfs_statfs(struct dentry *dentry, struct kstatfs *buf) {
  // TODO: ダミーデータではなく、本当の値を設定する。
  int total_blocks = 10000;
  int used_blocks = 1000;
  int total_files = 1000;
  int used_files = 50;

  struct kstatfs stat = {
      /* Type of filesystem */
      .f_type = ELTONFS_SUPER_MAGIC,
      /* Optimal transfer block size */
      .f_bsize = PAGE_SIZE,
      /* Total data blocks in filesystem */
      .f_blocks = total_blocks,
      /* Free blocks in filesystem */
      .f_bfree = total_blocks - used_blocks,
      /* Free blocks available to unprivileged user */
      .f_bavail = total_blocks - used_blocks,
      /* Total file nodes in filesystem */
      .f_files = total_files,
      /* Free file nodes in filesystem */
      .f_ffree = total_files - used_files,
      /* Filesystem ID */
      /* .f_fsid = ..., */
      /* Maximum length of filenames */
      .f_namelen = 100,
      /* Fragment size (since Linux 2.6) */
      /* .f_frsize = ..., */
      /* Mount flags of filesystem (since Linux 2.6.36) */
      /* .f_flags = , */
  };
  *buf = stat;
  return 0;
}

// Display the mount options in /proc/mounts.
static int eltonfs_show_options(struct seq_file *m, struct dentry *root) {
  // seq_puts(m, ",default");
  return 0;
}

static int __init fs_module_init(void) {
  int error;
  DEBUG("Loading the module ...");

#ifdef ELTONFS_UNIT_TEST
  INFO("Running unit test cases");
  test_xdr();
  test_rpc();
  if (IS_ASSERTION_FAILED()) {
    ERR("Failed to some unit test cases");
    GOTO_IF(out, -ENOTRECOVERABLE);
  }
  INFO("Finished all test cases without error");
#endif // ELTONFS_UNIT_TEST

  // Register filesystem.
  GOTO_IF(out, register_filesystem(&eltonfs_type));
  DEBUG("Registered eltonfs");

  // Start UMH and RPC server.
  GOTO_IF(out_unregister_fs, elton_rpc_server_init(&server, NULL));
  GOTO_IF(out_unregister_fs, server.ops->start_worker(&server));
  GOTO_IF(out_close_server, server.ops->start_umh(&server));
  DEBUG("Started an eltonfs user mode helper process");

  is_registered = 1;
  INFO("The module loaded");
  return 0;

out_close_server:
  server.ops->close(&server);
out_unregister_fs:
  unregister_filesystem(&eltonfs_type);
out:
  return error;
}

static void __exit fs_module_exit(void) {
  int error;
  DEBUG("Unloading the module ...");

  if (is_registered) {
    RETURN_VOID_IF(server.ops->close(&server));
    RETURN_VOID_IF(unregister_filesystem(&eltonfs_type));
  }

  INFO("The module unloaded");
}

int eltonfs_file_mmap(struct file *file, struct vm_area_struct *vma) {
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

static struct file_system_type eltonfs_type = {
    .owner = THIS_MODULE,
    .name = FS_NAME,
    .mount = mount,
    .kill_sb = kill_sb,
    .fs_flags = 0,
};
static struct super_operations eltonfs_s_op = {
    .alloc_inode = eltonfs_alloc_inode,
    .destroy_inode = eltonfs_destory_inode,
    .statfs = eltonfs_statfs,
    .drop_inode = generic_delete_inode,
    .show_options = eltonfs_show_options,
};
static struct address_space_operations eltonfs_aops = {
    .readpage = simple_readpage,
    .write_begin = simple_write_begin,
    .write_end = simple_write_end,
    .set_page_dirty = eltonfs_set_page_dirty,
};
static struct inode_operations eltonfs_file_inode_operations = {
    .setattr = simple_setattr,
    .getattr = simple_getattr,
#ifdef ELTONFS_XATTRS
    .listxattr = elton_list_xattr_vfs,
#endif
    .update_time = elton_update_time,
};
static struct inode_operations eltonfs_dir_inode_operations = {
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
static struct inode_operations eltonfs_symlink_inode_operations = {
    .get_link = page_get_link,
#ifdef ELTONFS_XATTRS
    .listxattr = elton_list_xattr_vfs,
#endif
};
static struct file_operations eltonfs_file_operations = {
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

module_init(fs_module_init);
module_exit(fs_module_exit);

MODULE_ALIAS_FS("elton");
MODULE_LICENSE("GPL v2");
MODULE_AUTHOR("yuuki0xff <yuuki0xff@gmail.com>");
MODULE_DESCRIPTION(MODULE_NAME " module");
