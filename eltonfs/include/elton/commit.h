#ifndef _ELTON_COMMIT_H
#define _ELTON_COMMIT_H

#include <elton/elton.h>
#include <elton/rpc/struct.h>

int get_commit_id_by_config(struct eltonfs_config *config, char **cid);
int get_commit_info(char *cid, struct commit_info **info);

struct tree_info *eltonfs_build_tree(struct inode *root);
char *eltonfs_call_commit(struct super_block *sb, struct tree_info *tree);
struct commit_info *eltonfs_get_commit(const char *cid);
void eltonfs_apply_tree(struct inode *inode, struct tree_info *tree);

#endif // _ELTON_COMMIT_H
