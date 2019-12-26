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

static inline bool eltonfs_is_valid_arg_name_char(char c) {
  return ('a' <= c && c <= 'z') || c == '_';
}
static int eltonfs_parse_arg_name(char **cursor, char **arg,
                                  bool *found_value) {
  char *start = *cursor;

  if (**cursor == '\0') {
    *arg = NULL;
    return 0;
  }

  // Seek an option.
  while (eltonfs_is_valid_arg_name_char(**cursor))
    (*cursor)++;

  // Check termination char.
  switch (**cursor) {
  case ',': // Found next argument.
    **cursor = '\0';
    (*cursor)++;
    // fallthrough
  case '\0': // Reached to end of string.
    *arg = start;
    *found_value = false;
    break;
  case '=': // Found a value.
    **cursor = '\0';
    (*cursor)++;
    *arg = start;
    *found_value = true;
    break;
  default: { // Invalid character found.
    char invalid = **cursor;
    (*cursor)++;
    **cursor = '\0';
    *arg = NULL;
    *found_value = false;
    ERR("invalid mount option \"%s\".  The character '0x%x' is not allowed in "
        "argument name of mount option.",
        start, invalid);
    return -EINVAL;
  }
  }
  return 0;
}
static bool eltonfs_is_valid_arg_value(char c) {
  return ('a' <= c && c <= 'z') || ('0' <= c && c <= '9') || c == '/';
}
static int eltonfs_parse_arg_value(char **cursor, char **value) {
  char *start = *cursor;

  if (**cursor == '\0') {
    *value = start;
    return 0;
  }

  // Seek a value.
  while (eltonfs_is_valid_arg_value(**cursor))
    (*cursor)++;

  // Check termination char.
  switch (**cursor) {
  case ',': // Found next argument.
    **cursor = '\0';
    (*cursor)++;
    // fallthrough
  case '\0': // Reached to end of string.
    *value = start;
    break;
  default: { // Invalid character found.
    char invalid = **cursor;
    (*cursor)++;
    **cursor = '\0';
    *value = NULL;
    ERR("invalid mount option \"%s\".  The character '0x%x' is not allowed in "
        "argument value of mount option.",
        start, invalid);
    return -EINVAL;
  }
  }
  return 0;
}
static int eltonfs_parse_opt(char *opt, struct eltonfs_config *config) {
  int error = 0;
  char *arg, *value;
  bool found_value;

  // Set default settings.
  config->auto_tx = true;
  config->vid = NULL;
  config->cid = NULL;
  config->vol_name = NULL;

  // Parse config string.
  while (1) {
    RETURN_IF(eltonfs_parse_arg_name(&opt, &arg, &found_value));
    if (arg == NULL) {
      // Reached to end of string.
      break;
    }

    if (!strcmp(arg, "vid")) {
      if (!found_value) {
        ERR("vid option requires an value");
        return -EINVAL;
      }
      RETURN_IF(eltonfs_parse_arg_value(&opt, &value));
      config->vid = kmalloc(strlen(value) + 1, GFP_KERNEL);
      if (config->vid == NULL) {
        return -ENOMEM;
      }
      strcpy(config->vid, value);
    } else if (!strcmp(arg, "cid")) {
      if (!found_value) {
        ERR("cid option requires an value");
        return -EINVAL;
      }
      RETURN_IF(eltonfs_parse_arg_value(&opt, &value));
      config->cid = kmalloc(strlen(value) + 1, GFP_KERNEL);
      if (config->cid == NULL) {
        return -ENOMEM;
      }
      strcpy(config->cid, value);
    } else if (!strcmp(arg, "vol")) {
      if (!found_value) {
        ERR("vol option requires an value");
        return -EINVAL;
      }
      RETURN_IF(eltonfs_parse_arg_value(&opt, &value));
      config->vol_name = kmalloc(strlen(value) + 1, GFP_KERNEL);
      if (config->vol_name == NULL) {
        return -ENOMEM;
      }
      strcpy(config->vol_name, value);
    } else if (!strcmp(arg, "auto_tx")) {
      if (found_value) {
        ERR("auto_tx is not accept a value");
        return -EINVAL;
      }
      config->auto_tx = true;
    } else if (!strcmp(arg, "no_auto_tx")) {
      if (found_value) {
        ERR("no_auto_tx is not accept a value");
        return -EINVAL;
      }
      config->auto_tx = false;
    } else {
      ERR("unrecognized mount option \"%s\"", arg);
      return -EINVAL;
    }
  }

  // Check arguments combination.
  {
    int counter = 0;
    if (config->vid)
      counter++;
    if (config->cid)
      counter++;
    if (config->vol_name)
      counter++;

    if (counter == 0) {
      ERR("require vid or cid or vol option");
      return -EINVAL;
    } else if (counter > 1) {
      ERR("vid and cid and vol options are exclusive");
      return -EINVAL;
    }
  }
  return 0;
}

static int eltonfs_fill_super(struct super_block *sb, void *data, int silent) {
  int error = 0;
  struct inode *inode;
  struct dentry *root;
  struct iattr ia;

  struct eltonfs_info *info = kmalloc(sizeof(struct eltonfs_info), GFP_KERNEL);
  RETURN_IF(eltonfs_parse_opt(data, &info->config));

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
static struct dentry *eltonfs_mount(struct file_system_type *fs_type, int flags,
                                    const char *dev_name, void *data) {
  return mount_nodev(fs_type, flags, data, eltonfs_fill_super);
}
static void kill_sb(struct super_block *sb) {
  // todo
}

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

#ifdef ELTONFS_UNIT_TEST
void test_super(void);
#endif // ELTONFS_UNIT_TEST

static int __init fs_module_init(void) {
  int error;
  DEBUG("Loading the module ...");

#ifdef ELTONFS_UNIT_TEST
  INFO("Running unit test cases");
  test_xdr();
  test_rpc();
  test_super();
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
    .mount = eltonfs_mount,
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

#ifdef ELTONFS_UNIT_TEST
static void test_eltonfs_parse_opt(void) {
  char opt[64];
  struct eltonfs_config config;
  // With vol.
  strcpy(opt, "vol=foo,no_auto_tx");
  if (!ASSERT_NO_ERROR(eltonfs_parse_opt(opt, &config))) {
    ASSERT_EQUAL_BOOL(false, config.auto_tx);
    ASSERT_EQUAL_BYTES("foo", config.vol_name, 4);
  }
  // With vid.
  strcpy(opt, "vid=34a50566000000");
  if (!ASSERT_NO_ERROR(eltonfs_parse_opt(opt, &config))) {
    ASSERT_EQUAL_BOOL(true, config.auto_tx);
    ASSERT_EQUAL_BYTES("34a50566000000", config.vid, 15);
  }
  // With cid.
  strcpy(opt, "cid=34a50566000000/14818143155257344");
  if (!ASSERT_NO_ERROR(eltonfs_parse_opt(opt, &config))) {
    ASSERT_EQUAL_BYTES("34a50566000000/14818143155257344", config.cid, 33);
  }
  // No argument.
  strcpy(opt, "");
  ASSERT_EQUAL_ERROR(-EINVAL, eltonfs_parse_opt(opt, &config));
  // Invalid argument name.
  strcpy(opt, "invalid");
  ASSERT_EQUAL_ERROR(-EINVAL, eltonfs_parse_opt(opt, &config));
  // Invalid value.
  strcpy(opt, "cid=###");
  ASSERT_EQUAL_ERROR(-EINVAL, eltonfs_parse_opt(opt, &config));
}
void test_super(void) { test_eltonfs_parse_opt(); }
#endif // ELTONFS_UNIT_TEST
