#ifndef _ELTON_XATTR_H
#define _ELTON_XATTR_H


#include <linux/xattr.h>

extern ssize_t elton_listxattr(struct dentry *dentry, char *buffer, size_t buffer_size);
extern const struct xattr_handler *elton_xattr_handlers[];


#endif // _ELTON_XATTR_H
