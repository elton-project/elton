#include "xattr.h"

extern ssize_t elton_listxattr(struct dentry *, char *, size_t) {
	// TODO
}

struct xattr_handler *elton_xattr_handlers[] = {
	NULL,
}

