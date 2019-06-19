#ifndef _ELTON_XATTR_H
#define _ELTON_XATTR_H


#include <linux/xattr.h>

extern ssize_t elton_listxattr(struct dentry *, char *, size_t);
extern const struct xattr_handler *elton_xattr_handlers[];


#endif // _ELTON_XATTR_H
