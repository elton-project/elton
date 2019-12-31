#include <elton/elton.h>

int eltonfs_create_cache_dir(void);
struct file *eltonfs_open_real_file(struct eltonfs_inode *inode,
                                    struct file *file);
