#include <elton/assert.h>
#include <elton/commit.h>
#include <elton/elton.h>
#include <elton/local_cache.h>
#include <elton/rpc/server.h>
#include <elton/rpc/test.h>
#include <elton/xattr.h>
#include <elton/xdr/test.h>
#include <linux/dcache.h>
#include <linux/fs.h>
#include <linux/kernel.h>
#include <linux/kthread.h>
#include <linux/module.h>
#include <linux/mount.h>
#include <linux/pagemap.h>
#include <linux/seq_file.h>
#include <linux/statfs.h>
#include <linux/wait.h>

static bool is_registered = 0;
struct elton_rpc_server server;
static struct file_system_type eltonfs_type;
static struct super_operations eltonfs_s_op;
struct dentry_operations eltonfs_dops;

int elton_update_time(struct inode *inode, struct timespec64 *time, int flags) {
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

long eltonfs_ioctl(struct file *file, unsigned int cmd, unsigned long arg) {
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
long eltonfs_compat_ioctl(struct file *file, unsigned int cmd,
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
  char *start;

  if (!*cursor) {
    // cursor is NULL.
    *arg = NULL;
    *found_value = false;
    return 0;
  }
  start = *cursor;

  if (**cursor == '\0') {
    // Empty string.
    *arg = NULL;
    *found_value = false;
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
static inline bool eltonfs_is_valid_arg_value(char c) {
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

struct fill_cid_args {
  // Arguments
  struct eltonfs_config *config;
  char **cid;
  struct commit_info **info;
  struct wait_queue_head *wq;
  struct spinlock *lock;

  // Results
  int error;
  bool finished;
};

static int _eltonfs_get_commit_info(void *_args) {
  int error = 0;
  char *cid = NULL;
  struct commit_info *info;
  struct fill_cid_args *args = (struct fill_cid_args *)_args;

  DEBUG("Getting cid from mount option and controll servers ...");
  GOTO_IF(out, get_commit_id_by_config(args->config, &cid));

  DEBUG("Getting initial data from controll servers ...");
  GOTO_IF(out, get_commit_info(cid, &info));

out:
  spin_lock(args->lock);
  *args->cid = cid;
  *args->info = info;
  args->error = error;
  args->finished = true;
  spin_unlock(args->lock);
  wake_up(args->wq);
  return 0;
}

static int eltonfs_fill_super(struct super_block *sb, void *data, int silent) {
  int error = 0;
  struct inode *inode = NULL;
  struct dentry *root;
  struct eltonfs_info *info = NULL;
  char *cid = NULL;
  struct commit_info *cinfo = NULL;

#ifdef ELTONFS_DEBUG
  DEBUG("%s: sb=%px, data=%px, silent=%d", __func__, sb, data, silent);
#endif // ELTONFS_DEBUG

  DEBUG("Creating cache directories");
  RETURN_IF(eltonfs_create_cache_dir());

  DEBUG("Initializing eltonfs_info");
  info = kzalloc(sizeof(struct eltonfs_info), GFP_KERNEL);
  if (!info)
    RETURN_IF(-ENOMEM);
  GOTO_IF(err, eltonfs_parse_opt(data, &info->config));
  info->cid = NULL;
  info->cred = get_current_cred();
  if (!info->cred)
    RETURN_IF(-ENOMEM);
  // todo: release cred if error occurred.
  info->inodes_vfs = kmalloc(sizeof(*info->inodes_vfs), GFP_KERNEL);
  if (!info->inodes_vfs)
    RETURN_IF(-ENOMEM);
  INIT_RADIX_TREE(info->inodes_vfs, GFP_NOFS);

  DEBUG("Getting commit information");
  {
    struct task_struct *task;
    struct wait_queue_head wq;
    struct spinlock lock;
    struct fill_cid_args fcargs = {
        .config = &info->config,
        .cid = &cid,
        .info = &cinfo,
        .wq = &wq,
        .lock = &lock,
    };
    init_waitqueue_head(&wq);
    spin_lock_init(&lock);
    task = kthread_run(_eltonfs_get_commit_info, &fcargs, "eltonfs-mount");
    spin_lock_irq(&lock);
    GOTO_IF(err, wait_event_interruptible_lock_irq(wq, fcargs.finished, lock));
    error = fcargs.error;
    info->cid = (const char *)cid;
    info->cinfo = cinfo;
    info->inodes_ei = cinfo->tree->inodes;
    spin_unlock_irq(&lock);
    GOTO_IF(err, error);
  }

  DEBUG("Preparing for super block ...");
  sb->s_blocksize_bits = PAGE_SHIFT;
  sb->s_blocksize = PAGE_SIZE;
  sb->s_maxbytes = MAX_LFS_FILESIZE;
  sb->s_type = &eltonfs_type;
  sb->s_op = &eltonfs_s_op;
  sb->s_d_op = &eltonfs_dops;
  sb->s_time_gran = 1;
  sb->s_fs_info = info;
#ifdef ELTONFS_XATTRS
  sb->s_xattr = elton_xattr_handlers;
#endif

  inode = vfs_i(eltonfs_iget(sb, info->cinfo->tree->root->eltonfs_ino));
  if (!inode) {
    WARN_ONCE(1, "root inode is null");
    GOTO_IF(err, -ENOMEM);
  }
  if (IS_ERR(inode))
    GOTO_IF(err, PTR_ERR(inode));

  root = d_make_root(inode);
  if (!root)
    GOTO_IF(err, -ENOMEM);
  sb->s_root = root;
  DEBUG("Prepared the super block");
  return 0;

err:
  if (info->inodes_vfs)
    kfree(info->inodes_vfs);
  if (info)
    kfree(info);
  if (inode)
    iput(inode);
  if (cid)
    kfree(cid);
  if (cinfo)
    elton_rpc_free_decoded_data(cinfo);
  return error;
}
static struct dentry *eltonfs_mount(struct file_system_type *fs_type, int flags,
                                    const char *dev_name, void *data) {
  return mount_nodev(fs_type, flags, data, eltonfs_fill_super);
}
static void eltonfs_kill_sb(struct super_block *sb) {
  // todo: impl
  // todo: release cred
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

  spin_lock_init(&i->lock);
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
  struct eltonfs_info *info = (struct eltonfs_info *)root->d_sb->s_fs_info;
  seq_puts(m, ",cid=");
  if (info->cid)
    seq_puts(m, info->cid);
  else {
    WARN_ONCE(1, "cid is null");
    seq_puts(m, "<null>");
  }
  return 0;
}

static struct dentry *eltonfs_d_real(struct dentry *dentry,
                                     const struct inode *inode) {
  if (inode && d_inode(dentry) == inode)
    // It is an eltonfs dentry.
    return dentry;

  if (!d_is_reg(dentry)) {
    WARN_ONCE(1, "real dentry not found");
    return dentry;
  }
  return eltonfs_get_real_dentry(eltonfs_i(d_inode(dentry)));
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

static struct file_system_type eltonfs_type = {
    .owner = THIS_MODULE,
    .name = FS_NAME,
    .mount = eltonfs_mount,
    .kill_sb = eltonfs_kill_sb,
    .fs_flags = 0,
};
static struct super_operations eltonfs_s_op = {
    .alloc_inode = eltonfs_alloc_inode,
    .destroy_inode = eltonfs_destory_inode,
    .statfs = eltonfs_statfs,
    .drop_inode = generic_delete_inode,
    .show_options = eltonfs_show_options,
};
struct address_space_operations eltonfs_aops = {};
struct dentry_operations eltonfs_dops = {
    .d_real = eltonfs_d_real,
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
  // With empty argument.
  strcpy(opt, "");
  ASSERT_EQUAL_ERROR(-EINVAL, eltonfs_parse_opt(opt, &config));
  // With no argument.
  ASSERT_EQUAL_ERROR(-EINVAL, eltonfs_parse_opt(NULL, &config));
  // Invalid argument name.
  strcpy(opt, "invalid");
  ASSERT_EQUAL_ERROR(-EINVAL, eltonfs_parse_opt(opt, &config));
  // Invalid value.
  strcpy(opt, "cid=###");
  ASSERT_EQUAL_ERROR(-EINVAL, eltonfs_parse_opt(opt, &config));
}
void test_super(void) { test_eltonfs_parse_opt(); }
#endif // ELTONFS_UNIT_TEST
