#include <elton/assert.h>
#include <elton/rpc/packet.h>
#include <elton/rpc/server.h>
#include <elton/rpc/struct.h>
#include <elton/xdr/error.h>
#include <elton/xdr/interface.h>
#include <linux/kthread.h>
#include <linux/un.h>
#include <net/sock.h>

#define LISTEN_LENGTH 4
#define READ_SOCK(sock, buff, size, offset)                                    \
  (sock)->file->f_op->read((sock)->file, (buff), (size), (offset));
#define WRITE_SOCK(sock, buff, size, offset)                                   \
  (sock)->file->f_op->read((sock)->file, (buff), (size), (offset));
// Get an entry from hashtable.  If entry is not found, returns NULL.
//
// Arguments:
//   hashtable:
//   type: the type name of entry.
//   member: the name of the hlist_node.
//   key_member: the field name that the hash key is stored.
//   key:
#define HASH_GET(hashtable, type, member, key_member, key)                     \
  ({                                                                           \
    type *obj = NULL;                                                          \
    hash_for_each_possible((hashtable), obj, member, (key)) {                  \
      if (obj->key_member == key)                                              \
        break;                                                                 \
    }                                                                          \
    obj;                                                                       \
  })
// Get session by session ID.  If session is not found, returns NULL.
// MUST NOT call when acquired the server->ss_table_lock.
#define GET_SESSION(server, session_id)                                        \
  ({                                                                           \
    struct elton_rpc_session *entry;                                           \
    spin_lock(&(server)->ss_table_lock);                                       \
    entry = HASH_GET((server)->ss_table, struct elton_rpc_session, _hash, sid, \
                     session_id);                                              \
    spin_unlock(&(server)->ss_table_lock);                                     \
    entry;                                                                     \
  })
// Add session to server.
// MUST NOT call when acquired the session->server->ss_table_lock.
#define ADD_SESSION(session)                                                   \
  do {                                                                         \
    spin_lock(&(session)->server->ss_table_lock);                              \
    hash_add((session)->server->ss_table, &(session)->_hash, (session)->sid);  \
    spin_unlock(&(session)->server->ss_table_lock);                            \
  } while (0);
// Delete session from server.
// MUST NOT call when acquired the session->server->ss_table_lock.
#define DELETE_SESSION(session)                                                \
  do {                                                                         \
    spin_lock(&(session)->server->ss_table_lock);                              \
    hash_del(&(session)->_hash);                                               \
    spin_unlock(&(session)->server->ss_table_lock);                            \
  } while (0);

static struct elton_rpc_setup2 setup2 = {
    .error = 0,
    .reason = "",
    .server_name = "eltonfs",
    .version_major = 0,    // todo
    .version_minor = 0,    // todo
    .version_revision = 0, // todo
};

static void elton_rpc_session_init(struct elton_rpc_session *s,
                                   struct elton_rpc_server *server, u8 sid,
                                   struct socket *sock) {
  s->server = server;
  s->sid = sid;
  s->sock = sock;
  s->task = NULL;
  mutex_init(&s->task_lock);
  elton_rpc_queue_init(&s->q);
}

// Handle an accepted session and start handshake process with client.
// If handshake is compleated, execute following tasks:
//   * Register it to available session list.
//   * Until close the socket, read continuously from the socket and enqueue it
//     to the receive packet queue.
//
// Arguments:
//   _s: pointer to struct elton_rpc_session.  MUST allocate s->sock with
//       kmalloc().  The s and s->sock will be close and release memory with
//       kfree() before this thread finished.
//
// Returns:
//   0:  Finished normally.
//   <0: Failed session worker with an error.
static int rpc_session_worker(void *_s) {
  int error = 0;
  struct elton_rpc_session *s = (struct elton_rpc_session *)_s;
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

      n = READ_SOCK(s->sock, buff, sizeof(buff), &readed);
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
      n = WRITE_SOCK(s->sock, raw->data, raw->size, &wrote);
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

  // todo: execute recv worker
  {
    struct raw_packet *raw = NULL;
    char *buff;
    size_t buff_size = PAGE_SIZE;
    loff_t readed = 0;
    size_t need_size;
    size_t consumed_size;

    BUILD_BUG_ON(sizeof(struct raw_packet) > PAGE_SIZE);

    buff = vmalloc(buff_size);
    if (buff == NULL) {
      error = -ENOMEM;
      goto error_alloc_buffer;
    }

    for (;;) {
      int n = READ_SOCK(s->sock, buff, sizeof(buff), &readed);
      if (n < 0) {
        error = n;
        goto error_read_sock;
      }

      error = elton_rpc_get_raw_packet_size(buff, buff_size, &need_size);
      if (error == -ELTON_XDR_NEED_MORE_MEM)
        // Need more bytes to calculate the raw packet size.
        continue;
      GOTO_IF(error_get_size, error); // Unexpected error.

      if (buff_size < need_size) {
        // Insufficient buffer size.  Increase buffer size.
        size_t new_size = round_up(need_size, PAGE_SIZE);
        char *new_buff = vmalloc(new_size);
        if (new_buff == NULL) {
          error = -ENOMEM;
          goto error_alloc_buffer;
        }
        memcpy(new_buff, buff, buff_size);
        vfree(buff);
        buff = new_buff;
        buff_size = new_size;
        continue;
      }
      if (readed < need_size)
        // Need more bytes to decode raw packet.
        continue;

      // Enough data was read to decode the raw packet.  Try to decode it.
      GOTO_IF(error_decode,
              elton_rpc_build_raw_packet(&raw, buff, readed, &consumed_size));
      // todo: enqueue raw.
      memmove(buff, buff + consumed_size, readed - consumed_size);
      readed -= consumed_size;
    }

  error_decode:
  error_get_size:
  error_read_sock:
  error_alloc_buffer:
    if (buff)
      vfree(buff);
    goto error_recv;
  }
error_recv:
error_setup1:
  kfree(s->sock);
  s->sock = NULL;
  // Unregister from s->server->ss_table.
  DELETE_SESSION(s);
  return error;
}

// Handle an listened socket and wait for the client to connect.
// When connected from client, accept it and start the rpc_session_worker().
//
// Arguments:
//   _srv: The pointer to struct elton_rpc_server.  MUST allocate srv->sock and
//         srv with kmalloc().  After this worker finished, the srv->sock is
//         closed and released by kfree().
//
// Returns:
//   0:  Finished normally.
//   <0: Failed master worker with an error.
static int rpc_master_worker(void *_srv) {
  int error = 0;
  u8 session_id;
  struct elton_rpc_server *srv = (struct elton_rpc_server *)_srv;

  for (session_id = 1;; session_id++) {
    struct elton_rpc_session *s = NULL;
    struct task_struct *task;
    struct socket *sock;

    if (session_id == 0)
      // Skips because the session ID is invalid.
      continue;
    if (GET_SESSION(srv, session_id))
      // Skips because this session ID already used.
      continue;
    // TODO: Detect session ID depletion.

    sock = kzalloc(sizeof(struct socket), GFP_KERNEL);
    if (sock == NULL) {
      error = -ENOMEM;
      goto error_accept;
    }
    s = kmalloc(sizeof(struct elton_rpc_session), GFP_KERNEL);
    if (s == NULL) {
      error = -ENOMEM;
      goto error_accept;
    }
    elton_rpc_session_init(s, srv, session_id, sock);

    GOTO_IF(error_accept, srv->sock->ops->accept(srv->sock, s->sock, 0, false));

    // Start session worker.
    task = (struct task_struct *)kthread_run(rpc_session_worker, s,
                                             "elton-rpc [%d]", session_id);
    if (IS_ERR(task)) {
      error = PTR_ERR(task);
      goto error_kthread;
    }
    mutex_lock(&s->task_lock);
    s->task = task;
    mutex_unlock(&s->task_lock);

    // Register new session.
    ADD_SESSION(s);
    continue;

  error_kthread:
    s->sock->ops->release(s->sock);
  error_accept:
    if (sock)
      kfree(sock);
    if (s)
      kfree(s);
    break;
  }
  return error;
}

// Listen unix domain socket and serve RPC in the background.
static int rpc_listen(struct elton_rpc_server *s) {
  int error = 0;
  struct sockaddr_un addr;
  struct task_struct *task;

  // Initialize socket and listen.
  GOTO_IF(error, sock_create(AF_UNIX, SOCK_STREAM, 0, &s->sock));
  addr.sun_family = AF_UNIX;
  strncpy(addr.sun_path, s->socket_path, UNIX_PATH_MAX);
  GOTO_IF(error_sock, s->sock->ops->bind(s->sock, (struct sockaddr *)&addr,
                                         sizeof(struct sockaddr_un)));
  GOTO_IF(error_sock, s->sock->ops->listen(s->sock, LISTEN_LENGTH));

  // Start master worker.
  task = kthread_run(rpc_master_worker, s, "elton-rpc [master]");
  if (IS_ERR(task)) {
    error = PTR_ERR(task);
    goto error_sock;
  }
  mutex_lock(&s->task_lock);
  s->task = task;
  mutex_unlock(&s->task_lock);
  return 0;

error_sock:
  sock_release(s->sock);
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
  server->sock = NULL;
  server->task = NULL;
  mutex_init(&server->task_lock);
  hash_init(server->ss_table);
  spin_lock_init(&server->ss_table_lock);
  hash_init(server->nss_table);
  spin_lock_init(&server->nss_table_lock);
  server->ops = &rpc_ops;
  return 0;
}
