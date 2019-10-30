#include <elton/assert.h>
#include <elton/rpc/packet.h>
#include <elton/rpc/server.h>
#include <elton/rpc/struct.h>
#include <linux/kthread.h>
#include <linux/un.h>
#include <net/sock.h>

#define LISTEN_LENGTH 4

struct worker_args {
  struct elton_rpc_server *server;
  struct socket *sock;
};
static struct elton_rpc_setup2 setup2 = {
    .error = 0,
    .reason = "",
    .server_name = "eltonfs",
    .version_major = 0,    // todo
    .version_minor = 0,    // todo
    .version_revision = 0, // todo
};

// Handle an accepted session and start handshake process with client.
// If handshake is compleated, execute following tasks:
//   * Register it to available session list.
//   * Until close the socket, read continuously from the socket and enqueue it
//     to the receive packet queue.
//
// Arguments:
//   _args: pointer to struct worker_args.  MUST allocate args and args->sock
//          with kmalloc().  The args and args->sock will be release memory with
//          kfree() before this thread finished.
//
// Returns:
//   0:  Finished normally.
//   <0: Failed session worker with an error.
static int rpc_session_worker(void *_args) {
  int error = 0;
  struct worker_args *args = (struct worker_args *)_args;
  struct socket *sock = args->sock;
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

error_setup1:
  kfree(sock);
  kfree(args);
  return error;
}

// Handle an listened socket and wait for the client to connect.
// When connected from client, accept it and start the rpc_session_worker().
//
// Arguments:
//   _args: The pointer to struct worker_args.  MUST close args->sock and
//          release args->sock and args.
//
// Returns:
//   0:  Finished normally.
//   <0: Failed master worker with an error.
static int rpc_master_worker(void *_args) {
  int error = 0;
  int worker_id;
  struct worker_args *args = (struct worker_args *)_args;
  struct socket *sock = args->sock;

  for (worker_id = 1;; worker_id++) {
    struct worker_args *newargs = NULL;
    struct socket *newsock = NULL;
    struct task_struct *task;

    newargs = kmalloc(sizeof(struct worker_args), GFP_KERNEL);
    if (newargs == NULL) {
      error = -ENOMEM;
      break;
    }
    newargs->server = args->server;
    newargs->sock = newsock = kzalloc(sizeof(struct socket), GFP_KERNEL);
    if (newsock == NULL) {
      error = -ENOMEM;
      break;
    }
    GOTO_IF(error_accept, sock->ops->accept(sock, newsock, 0, false));

    task = (struct task_struct *)kthread_run(rpc_session_worker, newargs,
                                             "elton-rpc [%d]", worker_id);
    if (IS_ERR(task)) {
      error = PTR_ERR(task);
      goto error_kthread;
    }
    continue;

  error_kthread:
    newsock->ops->release(newsock);
  error_accept:
    if (newsock)
      kfree(newsock);
    if (newargs)
      kfree(newargs);
    break;
  }

  kfree(args);
  return error;
}

// Listen unix domain socket and serve RPC in the background.
static int rpc_listen(struct elton_rpc_server *s) {
  int error = 0;
  struct socket *sock;
  struct sockaddr_un addr;
  struct task_struct *task;
  struct worker_args *args;

  // Initialize socket and listen.
  GOTO_IF(error, sock_create(AF_UNIX, SOCK_STREAM, 0, &sock));
  addr.sun_family = AF_UNIX;
  strncpy(addr.sun_path, s->socket_path, UNIX_PATH_MAX);
  GOTO_IF(error_sock, sock->ops->bind(sock, (struct sockaddr *)&addr,
                                      sizeof(struct sockaddr_un)));
  GOTO_IF(error_sock, sock->ops->listen(sock, LISTEN_LENGTH));

  // Initialize args.
  args = kmalloc(sizeof(struct worker_args), GFP_KERNEL);
  if (args == NULL) {
    error = -ENOMEM;
    goto error_sock;
  }
  args->server = s;
  args->sock = sock;

  // Start master worker.
  task = kthread_run(rpc_master_worker, args, "elton-rpc [master]");
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

// Start UMH (User Mode Helper) in the background.
int rpc_start_umh(struct elton_rpc_server *s) {
  char *argv[] = {
      ELTONFS_HELPER,
      "--socket",
      ELTONFS_HELPER_SOCK,
      NULL,
  };
  char *envp[] = {
      "HOME=/",
      "TERM=linux",
      "PATH=" PATH_ENV,
      NULL,
  };

  // todo: register subprocess_info to server.
  return call_usermodehelper(ELTONFS_HELPER, argv, envp, UMH_WAIT_EXEC);
}

// todo
static struct elton_rpc_operations rpc_ops = {
    .listen = rpc_listen,
    .start_umh = rpc_start_umh,
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
