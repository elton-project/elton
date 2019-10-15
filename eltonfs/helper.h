#include <linux/net.h>

struct eltonfs_helper {
    struct socket *sock;
};

static int eltonfs_start_helper(struct eltonfs_helper *helper);
static int eltonfs_stop_helper(struct eltonfs_helper *helper);
