#include <elton/elton.h>
#include <elton/utils.h>

int eltonfs_create_cache_dir(void);
struct file *eltonfs_open_real_file(struct eltonfs_inode *inode,
                                    struct file *file);
int eltonfs_cache_obj_async(struct eltonfs_job *job, const char *oid,
                            struct super_block *sb);
int eltonfs_cache_obj(const char *oid, struct super_block *sb);
