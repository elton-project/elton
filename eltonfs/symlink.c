#include <elton/elton.h>
#include <elton/xattr.h>

const char *eltonfs_get_link(struct dentry *dentry, struct inode *inode,
                             struct delayed_call *done) {
  return eltonfs_i(inode)->symlink.redirect_to;
}

struct inode_operations eltonfs_symlink_inode_operations = {
    .get_link = eltonfs_get_link,
#ifdef ELTONFS_XATTRS
    .listxattr = elton_list_xattr_vfs,
#endif
};
