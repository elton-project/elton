/*
  File: fs/xattr.c

  Extended attribute handling.

  Copyright (C) 2001 by Andreas Gruenbacher <a.gruenbacher@computer.org>
  Copyright (C) 2001 SGI - Silicon Graphics, Inc <linux-xfs@oss.sgi.com>
  Copyright (c) 2004 Red Hat, Inc., James Morris <jmorris@redhat.com>
  Copyright (c) 2019 Yuki HIRANO <yuuki0xff@gmail.com>
 */
#include <linux/fs.h>
#include <linux/xattr.h>

/*
 * Allocate new xattr and copy in the value; but leave the name to callers.
 */
struct simple_xattr *simple_xattr_alloc(const void *value, size_t size)
{
	struct simple_xattr *new_xattr;
	size_t len;

	/* wrap around? */
	len = sizeof(*new_xattr) + size;
	if (len < sizeof(*new_xattr))
		return NULL;

	new_xattr = kmalloc(len, GFP_KERNEL);
	if (!new_xattr)
		return NULL;

	new_xattr->size = size;
	memcpy(new_xattr->value, value, size);
	return new_xattr;
}

/*
 * xattr GET operation for in-memory/pseudo filesystems
 */
int simple_xattr_get(struct simple_xattrs *xattrs, const char *name,
		     void *buffer, size_t size)
{
	struct simple_xattr *xattr;
	int ret = -ENODATA;

	spin_lock(&xattrs->lock);
	list_for_each_entry(xattr, &xattrs->head, list) {
		if (strcmp(name, xattr->name))
			continue;

		ret = xattr->size;
		if (buffer) {
			if (size < xattr->size)
				ret = -ERANGE;
			else
				memcpy(buffer, xattr->value, xattr->size);
		}
		break;
	}
	spin_unlock(&xattrs->lock);
	return ret;
}

/**
 * simple_xattr_set - xattr SET operation for in-memory/pseudo filesystems
 * @xattrs: target simple_xattr list
 * @name: name of the extended attribute
 * @value: value of the xattr. If %NULL, will remove the attribute.
 * @size: size of the new xattr
 * @flags: %XATTR_{CREATE|REPLACE}
 *
 * %XATTR_CREATE is set, the xattr shouldn't exist already; otherwise fails
 * with -EEXIST.  If %XATTR_REPLACE is set, the xattr should exist;
 * otherwise, fails with -ENODATA.
 *
 * Returns 0 on success, -errno on failure.
 */
int simple_xattr_set(struct simple_xattrs *xattrs, const char *name,
		     const void *value, size_t size, int flags)
{
	struct simple_xattr *xattr;
	struct simple_xattr *new_xattr = NULL;
	int err = 0;

	/* value == NULL means remove */
	if (value) {
		new_xattr = simple_xattr_alloc(value, size);
		if (!new_xattr)
			return -ENOMEM;

		new_xattr->name = kstrdup(name, GFP_KERNEL);
		if (!new_xattr->name) {
			kfree(new_xattr);
			return -ENOMEM;
		}
	}

	spin_lock(&xattrs->lock);
	list_for_each_entry(xattr, &xattrs->head, list) {
		if (!strcmp(name, xattr->name)) {
			if (flags & XATTR_CREATE) {
				xattr = new_xattr;
				err = -EEXIST;
			} else if (new_xattr) {
				list_replace(&xattr->list, &new_xattr->list);
			} else {
				list_del(&xattr->list);
			}
			goto out;
		}
	}
	if (flags & XATTR_REPLACE) {
		xattr = new_xattr;
		err = -ENODATA;
	} else {
		list_add(&new_xattr->list, &xattrs->head);
		xattr = NULL;
	}
out:
	spin_unlock(&xattrs->lock);
	if (xattr) {
		kfree(xattr->name);
		kfree(xattr);
	}
	return err;

}

static bool xattr_is_trusted(const char *name)
{
	return !strncmp(name, XATTR_TRUSTED_PREFIX, XATTR_TRUSTED_PREFIX_LEN);
}

static int xattr_list_one(char **buffer, ssize_t *remaining_size,
			  const char *name)
{
	size_t len = strlen(name) + 1;
	if (*buffer) {
		if (*remaining_size < len)
			return -ERANGE;
		memcpy(*buffer, name, len);
		*buffer += len;
	}
	*remaining_size -= len;
	return 0;
}

/*
 * xattr LIST operation for in-memory/pseudo filesystems
 */
ssize_t simple_xattr_list(struct inode *inode, struct simple_xattrs *xattrs,
			  char *buffer, size_t size)
{
	bool trusted = capable(CAP_SYS_ADMIN);
	struct simple_xattr *xattr;
	ssize_t remaining_size = size;
	int err = 0;

#ifdef CONFIG_FS_POSIX_ACL
	if (IS_POSIXACL(inode)) {
		if (inode->i_acl) {
			err = xattr_list_one(&buffer, &remaining_size,
					     XATTR_NAME_POSIX_ACL_ACCESS);
			if (err)
				return err;
		}
		if (inode->i_default_acl) {
			err = xattr_list_one(&buffer, &remaining_size,
					     XATTR_NAME_POSIX_ACL_DEFAULT);
			if (err)
				return err;
		}
	}
#endif

	spin_lock(&xattrs->lock);
	list_for_each_entry(xattr, &xattrs->head, list) {
		/* skip "trusted." attributes for unprivileged callers */
		if (!trusted && xattr_is_trusted(xattr->name))
			continue;

		err = xattr_list_one(&buffer, &remaining_size, xattr->name);
		if (err)
			break;
	}
	spin_unlock(&xattrs->lock);

	return err ? err : size - remaining_size;
}
