#include <elton/assert.h>
#include <elton/rpc/packet.h>
#include <elton/rpc/server.h>
#include <elton/rpc/struct.h>
#include <linux/kthread.h>
#include <linux/un.h>
#include <net/sock.h>
#define LISTEN_LENGTH 4

static struct elton_rpc_setup2 setup2 = {
    .error = 0,
    .reason = "",
    .server_name = "eltonfs",
    .version_major = 0,    // todo
    .version_minor = 0,    // todo
    .version_revision = 0, // todo
};

static int rpc_session_worker(void *data) {
  int error = 0;
  struct socket *sock = (struct socket *)data;
  struct elton_rpc_setup1 *setup1;

  // Start handshake.
  {
    // Receiving setup1.
    char buff[50];
    struct raw_packet raw = {
        .struct_id = ELTON_RPC_SETUP1_ID,
        .data = buff,
    };
    ssize_t n;
    loff_t readed = 0;
    do {
      BUG_ON(readed >= sizeof(buff));

      n = sock->file->f_op->read(sock->file, buff, sizeof(buff), &readed);
      if (n < 0) {
        error = n;
        goto error_setup1;
      }
    } while (elton_rpc_decode_packet(&raw, (void **)&setup1));
  }

  INFO("rpc: connected from %s",
       setup1->client_name ? setup1->client_name : "no-name client");
  // todo: check setup1.

  elton_rpc_free_decoded_data(setup1);

  // Sending setup2.
  {
    struct packet pk = {
        .struct_id = ELTON_RPC_SETUP2_ID,
        .data = &setup2,
    };
    struct raw_packet *raw;
    ssize_t n;
    loff_t wrote = 0;

    // Encode data.
    RETURN_IF(elton_rpc_encode_packet(&pk, &raw));
    BUG_ON(raw == NULL);
    BUG_ON(raw->data == NULL);
    // Send data to client.
    while (wrote < raw->size) {
      n = sock->file->f_op->write(sock->file, raw->data, raw->size, &wrote);
      if (n < 0) {
        error = n;
        goto error_setup2;
      }
    }

  error_setup2:
    if (raw)
      raw->free(raw);
    if (error)
      goto error_setup1;
  }

  // todo: register new session.
  // todo: execute recv worker
  return 0;

error_setup1:
  return error;
}

static int rpc_master_worker(void *data) {
  int error = 0;
  int worker_id;
  struct socket *sock = (struct socket *)data;

  for (worker_id = 1;; worker_id++) {
    struct socket *newsock;
    struct task_struct *task;

    newsock = kzalloc(sizeof(struct socket), GFP_KERNEL);
    if (newsock == NULL) {
      error = -ENOMEM;
      break;
    }
    GOTO_IF(error_accept, sock->ops->accept(sock, newsock, 0, false));

    task = (struct task_struct *)kthread_run(rpc_session_worker, newsock,
                                             "elton-rpc [%d]", worker_id);
    if (IS_ERR(task)) {
      error = PTR_ERR(task);
      goto error_kthread;
    }
    continue;

  error_kthread:
    newsock->ops->release(newsock);
  error_accept:
    kfree(newsock);
    break;
  }
  return error;
}

static int rpc_listen(struct elton_rpc_server *s) {
  int error = 0;
  struct socket *sock;
  struct sockaddr_un addr;
  struct task_struct *task;

  GOTO_IF(error, sock_create(AF_UNIX, SOCK_STREAM, 0, &sock));
  addr.sun_family = AF_UNIX;
  strncpy(addr.sun_path, s->socket_path, UNIX_PATH_MAX);

  GOTO_IF(error_sock, sock->ops->bind(sock, (struct sockaddr *)&addr,
                                      sizeof(struct sockaddr_un)));
  GOTO_IF(error_sock, sock->ops->listen(sock, LISTEN_LENGTH));

  task = kthread_run(rpc_master_worker, sock, "elton-rpc [master]");
  if (IS_ERR(task)) {
    error = PTR_ERR(task);
    goto error_sock;
  }
  return 0;

error_sock:
  sock_release(sock);
error:
  return error;
}

// todo
static struct elton_rpc_operations rpc_ops = {
    .listen = rpc_listen,
    .start_umh = NULL,
    .new_session = NULL,
    .close_nowait = NULL,
    .close = NULL,
};

int elton_rpc_server_init(struct elton_rpc_server *server, char *socket_path) {
  if (socket_path == NULL)
    socket_path = ELTON_UMH_SOCK;
  server->socket_path = socket_path;
  server->ops = &rpc_ops;
  return 0;
}
