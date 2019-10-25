#ifndef _ELTON_HELPER_H
#define _ELTON_HELPER_H

#include <linux/net.h>

struct eltonfs_helper {
    struct socket *sock;
};

int eltonfs_start_helper(struct eltonfs_helper *helper);
int eltonfs_stop_helper(struct eltonfs_helper *helper);

#endif // _ELTON_HELPER_H
