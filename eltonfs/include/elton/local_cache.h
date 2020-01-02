#ifndef _ELTON_LOCAL_CACHE_H
#define _ELTON_LOCAL_CACHE_H

#include <elton/elton.h>
#include <elton/utils.h>

#define REAL_PATH_MAX 120
#define CACHE_ID_LENGTH 32

#define PREFIX_CACHE_DIR "/var/cache/eltonfs"
#define PREFIX_LIB_DIR "/var/lib/eltonfs"

// リモートIDを持つオブジェクトを保存するディレクトリ
#define REMOTE_OBJ_DIR PREFIX_CACHE_DIR "/remote-objects"
// ローカルIDを持つオブジェクトを保存するディレクトリ
#define LOCAL_OBJ_DIR PREFIX_LIB_DIR "/local-objects"

int eltonfs_generate_cache_id(const char *base_dir, char fpath[REAL_PATH_MAX],
                              char id[CACHE_ID_LENGTH], struct inode **inode);
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
