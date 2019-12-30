#include <elton/elton.h>
#include <elton/xattr.h>

struct inode_operations eltonfs_symlink_inode_operations = {
    .get_link = page_get_link,
#ifdef ELTONFS_XATTRS
    .listxattr = elton_list_xattr_vfs,
#endif
};
