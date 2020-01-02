#ifndef _ELTON_LOCAL_CACHE_H
#define _ELTON_LOCAL_CACHE_H

#include <elton/elton.h>
#include <elton/utils.h>

int eltonfs_create_cache_dir(void);
struct file *eltonfs_open_real_file(struct eltonfs_inode *inode,
                                    struct file *file);
int eltonfs_cache_obj_async(struct eltonfs_job *job, const char *oid,
                            struct super_block *sb);
int eltonfs_cache_obj(const char *oid, struct super_block *sb);

#define OBJ_CACHE_ACCESS_START(_super_block)                                   \
  const struct cred *__old_cred = override_creds(eltonfs_sb(_super_block)->cred)
#define OBJ_CACHE_ACCESS_START_FILE(_file)                                     \
  OBJ_CACHE_ACCESS_START((_file)->f_inode->i_sb)

#define OBJ_CACHE_ACCESS_END revert_creds(__old_cred)

#endif // _ELTON_LOCAL_CACHE_H
