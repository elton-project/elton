#ifndef _ELTON_COMMIT_H
#define _ELTON_COMMIT_H

#include <elton/elton.h>
#include <elton/rpc/struct.h>

int get_commit_id_by_config(struct eltonfs_config *config, char **cid);
int get_commit_info(char *cid, struct commit_info **info);

#endif // _ELTON_COMMIT_H
