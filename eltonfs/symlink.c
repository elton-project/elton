#include <elton/elton.h>
#include <elton/local_cache.h>
#include <elton/xattr.h>

static inline void UPDATE_SIZE(struct inode *inode, size_t len) {
  i_size_write(inode, len);
  WRITE_ONCE(inode->i_blocks, len / i_blocksize(inode));
}

// Try to load symlink data from remote.
static inline int maybe_load_symlink(struct inode *inode) {
  int error = 0;
  const char *p;
  struct eltonfs_inode *ei = eltonfs_i(inode);
  if (likely(ei->symlink.redirect_to))
    // Fast path.
    return 0;

  // Slow path.
  if (WARN_ON_ONCE(!ei->symlink.object_id))
    return -EINVAL;
  error = eltonfs_cache_obj(ei->symlink.object_id, inode->i_sb);
  if (error)
    return error;
  p = eltonfs_read_obj(ei->symlink.object_id, inode->i_sb);
  if (IS_ERR(p))
    return PTR_ERR(p);
  ei->symlink.redirect_to = p;
  UPDATE_SIZE(inode, strlen(p));
  return 0;
}

static const char *eltonfs_get_link(struct dentry *dentry, struct inode *inode,
                                    struct delayed_call *done) {
  int error = maybe_load_symlink(inode);
  if (error)
    return ERR_PTR(error);
  return eltonfs_i(inode)->symlink.redirect_to;
}

static int eltonfs_symlink_getattr(const struct path *path, struct kstat *stat,
                                   u32 request_mask, unsigned int query_flags) {
  struct inode *inode = d_inode(path->dentry);
  int error = maybe_load_symlink(inode);
  if (error)
    return error;
  generic_fillattr(inode, stat);
  return 0;
}

const struct inode_operations eltonfs_symlink_inode_operations = {
    .get_link = eltonfs_get_link,
    .getattr = eltonfs_symlink_getattr,
#ifdef ELTONFS_XATTRS
    .listxattr = elton_list_xattr_vfs,
#endif
};
