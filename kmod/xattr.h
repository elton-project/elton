#ifndef _ELTON_XATTR_H
#define _ELTON_XATTR_H


#include <linux/xattr.h>

ssize_t elton_list_xattr(struct inode *inode, char *buffer, size_t buffer_size);
int elton_set_xattr(struct inode *inode, char *name, void *value, size_t size, int flags);
int elton_get_xattr(struct inode *inode, char *name, void *value, size_t size);

extern struct xattr_handler *elton_xattr_handlers[];


#endif // _ELTON_XATTR_H
