#ifdef ELTONFS_XATTRS

#include <elton/assert.h>
#include <elton/elton.h>
#include <elton/xattr.h>

ssize_t elton_list_xattr_vfs(struct dentry *dentry, char *buffer,
                             size_t buffer_size) {
  RETURN_EXTERNAL(elton_list_xattr(d_inode(dentry), buffer, buffer_size));
}
ssize_t elton_list_xattr(struct inode *inode, char *buffer,
                         size_t buffer_size) {
  struct eltonfs_inode *i = eltonfs_i(inode);
  return simple_xattr_list(inode, &i->xattrs, buffer, buffer_size);
}

int elton_set_xattr(struct inode *inode, const char *name, const void *value,
                    size_t size, int flags) {
  struct eltonfs_inode *i = eltonfs_i(inode);
  return simple_xattr_set(&i->xattrs, name, value, size, flags);
}

int elton_get_xattr(struct inode *inode, const char *name, void *value,
                    size_t size) {
  struct eltonfs_inode *i = eltonfs_i(inode);
  return simple_xattr_get(&i->xattrs, name, value, size);
}

static int elton_xattr_handler_set(const struct xattr_handler *handler,
                                   struct dentry *dentry, struct inode *inode,
                                   const char *name, const void *buffer,
                                   size_t size, int flags) {
  const char *full_name = xattr_full_name(handler, name);
  return elton_set_xattr(inode, full_name, buffer, size, flags);
}

static int elton_xattr_handler_get(const struct xattr_handler *handler,
                                   struct dentry *dentry, struct inode *inode,
                                   const char *name, void *buffer,
                                   size_t size) {
  const char *full_name = xattr_full_name(handler, name);
  return elton_get_xattr(inode, full_name, buffer, size);
}

const static struct xattr_handler elton_xattr_user_handler = {
    .prefix = XATTR_USER_PREFIX,
    .get = elton_xattr_handler_get,
    .set = elton_xattr_handler_set};
const static struct xattr_handler elton_xattr_system_handler = {
    .prefix = XATTR_SYSTEM_PREFIX,
    .get = elton_xattr_handler_get,
    .set = elton_xattr_handler_set};
const static struct xattr_handler elton_xattr_security_handler = {
    .prefix = XATTR_SECURITY_PREFIX,
    .get = elton_xattr_handler_get,
    .set = elton_xattr_handler_set};
const static struct xattr_handler elton_xattr_trusted_handler = {
    .prefix = XATTR_TRUSTED_PREFIX,
    .get = elton_xattr_handler_get,
    .set = elton_xattr_handler_set};

const struct xattr_handler *elton_xattr_handlers[] = {
    &elton_xattr_user_handler,
    &elton_xattr_system_handler,
    &elton_xattr_security_handler,
    &elton_xattr_trusted_handler,
    NULL,
};

#endif // ELTONFS_XATTRS
