#include <linux/umh.h>
#include <linux/net.h>
#include <uapi/linux/un.h>
#include "elton.h"
#include "assert.h"

static int eltonfs_serve_server(struct socket **sockp) {
    int error;
    struct socket *sock;
    struct sockaddr_un addr;

    // Create socket
    error = sock_create(AF_UNIX, SOCK_STREAM, 0, sockp);
    if(CHECK_ERROR(error)) {
        ERR("Failed to create socket");
        goto out;
    }
    sock = *sockp;

    // Bind
    memset(&addr, 0, sizeof(addr));
    addr.sun_family = AF_UNIX;
    strncpy(addr.sun_path, ELTONFS_HELPER_SOCK, UNIX_PATH_MAX);
    error = sock->ops->bind(sock, (struct sockaddr*)&addr, sizeof(addr));
    if(CHECK_ERROR(error)) {
        ERR("Failed to bind socket");
        goto out_close_socket;
    }

    // Listen
    error = sock->ops->listen(sock, 10);
    if(CHECK_ERROR(error)) {
        ERR("Failed to listen socket");
        goto out_close_socket;
    }
    return 0;

out_close_socket:
    sock->ops->release(sock);  // Ignore error.
out:
    return error;
}

static int eltonfs_stop_server(struct socket **sockp) {
    int error;
    struct socket *sock = *sockp;

    error = sock->ops->release(sock);
    if(CHECK_ERROR(error)){
        ERR("Failed to close socket");
        return error;
    }
    return 0;
}

static int eltonfs_start_helper(struct socket **sockp) {
    int error;
    char *argv[] = {
        ELTONFS_HELPER,
        NULL,
    };
    char *envp[] = {
        ELTONFS_HELPER,
        NULL,
    };

    error = eltonfs_serve_server(sockp);
    if(CHECK_ERROR(error)) {
        goto out;
    }

    error = call_usermodehelper(
        "eltonfs-helper",
        argv, envp,
        UMH_WAIT_EXEC);
    if(CHECK_ERROR(error)) {
        goto out_stop_server;
    }
    return 0;

out_stop_server:
    eltonfs_stop_server(sockp);
out:
    return error;
}

static int eltonfs_stop_helper(struct socket **sockp) {
    return eltonfs_stop_server(sockp);
}
