#include <elton/assert.h>
#include <elton/rpc/error.h>
#include <elton/rpc/packet.h>
#include <elton/rpc/server.h>
#include <elton/rpc/struct.h>
#include <elton/xdr/error.h>
#include <elton/xdr/interface.h>
#include <linux/kthread.h>
#include <linux/syscalls.h>
#include <linux/un.h>
#include <net/sock.h>

#define __RPC_LOG_PREFIX "RPC: "
#define __SESSION_LOG_PREFIX "RPC sid=%d: "
#define __NS_LOG_PREFIX "RPC nsid=%llu: "

#ifdef ELTON_RPC_DEBUG
#define RPC_DEBUG(fmt, ...) DEBUG(__RPC_LOG_PREFIX fmt, ##__VA_ARGS__)
#define SESSION_DEBUG(session, fmt, ...)                                       \
  DEBUG(__SESSION_LOG_PREFIX fmt, session->sid, ##__VA_ARGS__)
#define NS_DEBUG(ns, fmt, ...)                                                 \
  DEBUG(__NS_LOG_PREFIX fmt, ns->nsid, ##__VA_ARGS__)
#else
#define RPC_DEBUG(fmt, ...)
#define SESSION_DEBUG(session, fmt, ...)
#define NS_DEBUG(ns, fmt, ...)
#endif

#define RPC_INFO(fmt, ...) INFO(__RPC_LOG_PREFIX fmt, ##__VA_ARGS__)
#define SESSION_INFO(session, fmt, ...)                                        \
  INFO(__SESSION_LOG_PREFIX fmt, session->sid, ##__VA_ARGS__)
#define NS_INFO(ns, fmt, ...) INFO(__NS_LOG_PREFIX fmt, ns->nsid, ##__VA_ARGS__)
#define RPC_ERR(fmt, ...) ERR(__RPC_LOG_PREFIX fmt, ##__VA_ARGS__)
#define SESSION_ERR(session, fmt, ...)                                         \
  ERR(__SESSION_LOG_PREFIX fmt, session->sid, ##__VA_ARGS__)
#define NS_ERR(ns, fmt, ...) ERR(__NS_LOG_PREFIX fmt, ns->nsid, ##__VA_ARGS__)

#define LISTEN_LENGTH 4
#define READ_SOCK(sock, buff, size, offset)                                    \
  ({                                                                           \
    struct iovec iov = {                                                       \
        .iov_base = (buff) + (*offset),                                        \
        .iov_len = (size) - (*offset),                                         \
    };                                                                         \
    struct iov_iter iter;                                                      \
    struct kiocb kiocb;                                                        \
    ssize_t result;                                                            \
    BUG_ON((sock)->file == NULL);                                              \
    iov_iter_init(&iter, READ, &iov, 1, iov.iov_len);                          \
    init_sync_kiocb(&kiocb, (sock)->file);                                     \
    kiocb.ki_pos = 0;                                                          \
    result = (sock)->file->f_op->read_iter(&kiocb, &iter);                     \
    if (result > 0)                                                            \
      *(offset) += result;                                                     \
    result;                                                                    \
  })
#define WRITE_SOCK(sock, buff, size, offset)                                   \
  ({                                                                           \
    struct iovec iov = {                                                       \
        .iov_base = (buff) + (*offset),                                        \
        .iov_len = (size) - (*offset),                                         \
    };                                                                         \
    struct iov_iter iter;                                                      \
    struct kiocb kiocb;                                                        \
    ssize_t result;                                                            \
    BUG_ON((sock)->file == NULL);                                              \
    iov_iter_init(&iter, WRITE, &iov, 1, iov.iov_len);                         \
    init_sync_kiocb(&kiocb, (sock)->file);                                     \
    kiocb.ki_pos = 0;                                                          \
    result = (sock)->file->f_op->write_iter(&kiocb, &iter);                    \
    if (result > 0)                                                            \
      *(offset) += result;                                                     \
    result;                                                                    \
  })
#define SOCK_PROTO_NAME(sock) (sock)->sk->sk_prot_creator->name

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
  } while (0)
// Delete session from server.
// MUST NOT call when acquired the session->server->ss_table_lock.
#define DELETE_SESSION(session)                                                \
  do {                                                                         \
    spin_lock(&(session)->server->ss_table_lock);                              \
    hash_del(&(session)->_hash);                                               \
    spin_unlock(&(session)->server->ss_table_lock);                            \
  } while (0)
// Loops all sessions in the server->ss_table.
// MUST NOT call when acquired the server->ss_table_lock.
#define FOR_EACH_SESSIONS(server, session, code_block)                         \
  do {                                                                         \
    int i;                                                                     \
    spin_lock(&(server)->ss_table_lock);                                       \
    hash_for_each((server)->ss_table, i, (session), _hash) code_block;         \
    spin_unlock(&(server)->ss_table_lock);                                     \
  } while (0)
// Add nested session to server.
// MUST NOT call when acquired the server->nss_table_lock.
#define ADD_NS(ns)                                                             \
  do {                                                                         \
    spin_lock(&(ns)->session->server->nss_table_lock);                         \
    hash_add((ns)->session->server->nss_table, &(ns)->_hash, (ns)->nsid);      \
    spin_unlock(&(ns)->session->server->nss_table_lock);                       \
  } while (0)
#define GET_NS_BY_HASH(server, hash)                                           \
  ({                                                                           \
    struct elton_rpc_ns *entry;                                                \
    spin_lock(&(server)->nss_table_lock);                                      \
    entry =                                                                    \
        HASH_GET((server)->nss_table, struct elton_rpc_ns, _hash, nsid, hash); \
    spin_unlock(&(server)->nss_table_lock);                                    \
    entry;                                                                     \
  })

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
  s->sock_wb = NULL;
  s->sock_wb_size = 0;
  s->task = NULL;
  mutex_init(&s->task_lock);
}

static void elton_rpc_ns_init(struct elton_rpc_ns *ns,
                              struct elton_rpc_session *s, u64 nsid,
                              bool is_client);

static u64 get_nsid_hash_by_values(u8 session_id, u64 nsid);

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

    SESSION_DEBUG(s, "waiting setup1 ...");
    do {
      BUG_ON(readed >= sizeof(buff));

      n = READ_SOCK(s->sock, buff, sizeof(buff), &readed);
      if (n < 0) {
        error = n;
        SESSION_ERR(s, "failed handshake on setup1 stage: read error %d",
                    error);
        goto error_setup1;
      }
    } while (elton_rpc_decode_packet(&raw, (void **)&setup1));
    SESSION_DEBUG(s, "received setup1 from client");
  }

  SESSION_DEBUG(s, "validating setup1");
  // todo: check setup1.
  SESSION_INFO(s, "connected from %s",
               setup1->client_name ? setup1->client_name : "no-name client");
  elton_rpc_free_decoded_data(setup1);

  // Sending setup2.
  {
    struct packet pk = {
        .struct_id = ELTON_RPC_SETUP2_ID,
        .data = &setup2,
    };
    struct raw_packet *raw = NULL;
    ssize_t n;
    loff_t wrote = 0;

    // Encode data.
    SESSION_DEBUG(s, "preparing setup2 ...");
    GOTO_IF(error_setup2, elton_rpc_encode_packet(&pk, &raw));
    BUG_ON(raw == NULL);
    BUG_ON(raw->data == NULL);
    // Send data to client.
    SESSION_DEBUG(s, "sending setup2 ...");
    while (wrote < raw->size) {
      n = WRITE_SOCK(s->sock, raw->data, raw->size, &wrote);
      if (n < 0) {
        error = n;
        SESSION_ERR(s, "failed handshake on setup2 stage: write error %d",
                    error);
        goto error_setup2;
      }
    }
    SESSION_DEBUG(s, "sent setup2");

  error_setup2:
    if (raw)
      raw->free(raw);
    if (error)
      goto error_setup1;
  }
  SESSION_INFO(s, "established connection");

  // Receive data from client until socket is closed.
  {
    struct raw_packet *raw = NULL;
    char *buff;
    size_t buff_size = PAGE_SIZE;
    loff_t readed = 0;
    size_t need_size;
    size_t consumed_size;
    u64 nsid_hash;
    struct elton_rpc_ns *ns = NULL;

    BUILD_BUG_ON(sizeof(struct raw_packet) > PAGE_SIZE);

    buff = vmalloc(buff_size);
    if (buff == NULL) {
      error = -ENOMEM;
      SESSION_ERR(s, "failed to allocate %zu bytes for read buffer", buff_size);
      goto error_alloc_buffer;
    }

    for (;;) {
      int n = READ_SOCK(s->sock, buff, sizeof(buff), &readed);
      if (n < 0) {
        error = n;
        SESSION_ERR(s, "read error %d", error);
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
          SESSION_ERR(s, "failed to allocate %zu bytes for read buffer",
                      new_size);
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
      memmove(buff, buff + consumed_size, readed - consumed_size);
      readed -= consumed_size;

      // Qneueue raw packet.
      spin_lock(&s->server->nss_table_lock);
      nsid_hash = get_nsid_hash_by_values(s->sid, raw->session_id);
      ns = GET_NS_BY_HASH(s->server, nsid_hash);
      SESSION_DEBUG(s, "received a packet: struct_id=%llu, flags=%d, size=%zu",
                    raw->struct_id, raw->flags, raw->size);
      if (ns) {
        // Enqueue it.
        elton_rpc_enqueue(&ns->q, raw);
      } else if (raw->flags & ELTON_SESSION_FLAG_CREATE) {
        // Create session and enqueue it.
        ns = (struct elton_rpc_ns *)kmalloc(sizeof(struct elton_rpc_ns),
                                            GFP_KERNEL);
        if (ns == NULL) {
          error = -ENOMEM;
          SESSION_ERR(s, "failed to allocate elton_rpc_ns object");
          goto error_enqueue;
        }
        elton_rpc_ns_init(ns, s, raw->session_id, false);
        SESSION_DEBUG(s, "created new session by umh");
        ADD_NS(ns);
        ns = NULL;
      }

    error_enqueue:
      spin_unlock(&s->server->nss_table_lock);
      if (raw)
        raw->free(raw);
      raw = NULL;

      if (error)
        break;
    }

  error_decode:
  error_get_size:
  error_read_sock:
  error_alloc_buffer:
    SESSION_INFO(s, "stopping receiver");
    if (buff)
      vfree(buff);
    goto error_recv;
  }
error_recv:
error_setup1:
  SESSION_INFO(s, "stopping session worker");
  kfree(s->sock);
  s->sock = NULL;
  // Unregister from s->server->ss_table.
  DELETE_SESSION(s);
  return error;
}

int rpc_sock_set_file(struct socket *newsock, char *name) {
  int error;
  int newfd;
  struct file *newfile;

  newfd = get_unused_fd_flags(0);
  if (newfd < 0) {
    error = newfd;
    goto error_newfd;
  }
  newfile = sock_alloc_file(newsock, 0, name);
  if (IS_ERR(newfile)) {
    error = PTR_ERR(newfile);
    goto error_alloc_file;
  }
  fd_install(newfd, newfile);
  return 0;

error_alloc_file:
  put_unused_fd(newfd);
error_newfd:
  return error;
}

int rpc_sock_accpet(struct socket *sock, struct socket **new) {
  int error;
  struct socket *newsock;

  newsock = sock_alloc();
  if (!newsock) {
    error = PTR_ERR(newsock);
    goto error_sock_alloc;
  }
  newsock->type = sock->type;
  newsock->ops = sock->ops;
  GOTO_IF(error_set_file, rpc_sock_set_file(newsock, SOCK_PROTO_NAME(sock)));

  error = sock->ops->accept(sock, newsock, sock->file->f_flags, false);
  if (error < 0)
    goto error_accept;

  *new = newsock;
  return error;

error_accept:
error_set_file:
  sock_release(newsock);
error_sock_alloc:
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
    struct socket *newsock = NULL;

    if (session_id == 0)
      // Skips because the session ID is invalid.
      continue;
    if (GET_SESSION(srv, session_id))
      // Skips because this session ID already used.
      continue;
    // TODO: Detect session ID depletion.

    GOTO_IF(error_accept, rpc_sock_accpet(srv->sock, &newsock));

    s = kmalloc(sizeof(struct elton_rpc_session), GFP_KERNEL);
    if (s == NULL) {
      error = -ENOMEM;
      RPC_ERR("master: failed to allocate elton_rpc_session");
      goto error_accept;
    }
    elton_rpc_session_init(s, srv, session_id, newsock);

    // Start session worker.
    task = (struct task_struct *)kthread_run(rpc_session_worker, s,
                                             "elton-rpc [%d]", session_id);
    if (IS_ERR(task)) {
      error = PTR_ERR(task);
      RPC_ERR("master: kthread_run returns an error %d", error);
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
    if (newsock)
      sock_release(newsock);
    if (s)
      kfree(s);
    break;
  }
  RPC_ERR("master: stopped");
  return error;
}

// Change the file mode.
static int rpc_sock_chmod(const char *path, umode_t mode) {
  return do_fchmodat(AT_FDCWD, path, mode);
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
  GOTO_IF(error_sock, rpc_sock_set_file(s->sock, SOCK_PROTO_NAME(s->sock)));
  GOTO_IF(error_sock, s->sock->ops->bind(s->sock, (struct sockaddr *)&addr,
                                         sizeof(struct sockaddr_un)));
  // Correct socket file mode.  We can safely change the mode of socket file
  // because it is not listened yet.
  GOTO_IF(error_sock, rpc_sock_chmod(s->socket_path, 0600));
  GOTO_IF(error_sock, s->sock->ops->listen(s->sock, LISTEN_LENGTH));
  RPC_DEBUG("listened UNIX Domain Socket");

  // Start master worker.
  task = kthread_run(rpc_master_worker, s, "elton-rpc [master]");
  if (IS_ERR(task)) {
    error = PTR_ERR(task);
    RPC_ERR("kthread_run returns an error %d", error);
    goto error_sock;
  }
  RPC_INFO("started master worker");
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
static int rpc_start_umh(struct elton_rpc_server *s) {
  int error;
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
  error = call_usermodehelper(ELTONFS_HELPER, argv, envp, UMH_WAIT_EXEC);
  if (error) {
    RPC_ERR("failed to start UMH with error %d", error);
    return error;
  }
  RPC_INFO("start " ELTONFS_HELPER);
  return error;
}

// Get new nsid.
//
// ============== Bit field structure of nsid ==============
//
// 63                          32                          0
// .      .      .      .      .      .      .      .      .
// |           32bit           | |         31bit           |
// |      reserved area        | |        sequence         |
//                              ^
//                              |
//                             1bit
//                          Client flag
static u64 get_nsid(void) {
  static DEFINE_SPINLOCK(lock);
  static u64 sequence = 0;
  u64 i;
  const u64 SEQUENCE_MASK = (1U << 31) - 1U;

  spin_lock(&lock);
  sequence++;
  i = sequence;
  spin_unlock(&lock);

  return i & SEQUENCE_MASK;
}

static u64 get_nsid_hash_by_values(u8 session_id, u64 nsid) {
  const int SID_SHIFT = 32;
  return ((u64)session_id) << SID_SHIFT | nsid;
}

// Get hash value of specified NS.
//
// ============ Bit field structure of nsid hash ===========
//
// 63                          32                          0
// .      .      .      .      .      .      .      .      .
// |        24bit       | 8bit | |         31bit           |
// |   reserved area    | sid  | |        sequence         |
//                              ^
//                              |
//                             1bit
//                          Client flag
static u64 get_nsid_hash(struct elton_rpc_ns *ns) {
  return get_nsid_hash_by_values(ns->session->sid, ns->nsid);
}

static int rpc_new_session(struct elton_rpc_server *srv,
                           struct elton_rpc_ns *ns) {
  struct elton_rpc_session *s = NULL;
  u64 nsid;
  u64 hash;
  bool found;

  // Select a session.
  FOR_EACH_SESSIONS(srv, s, break);
  if (s == NULL)
    // Session not found.  Elton needs at least one session to create nested
    // session.
    return -EINVAL;

  // Find the unused nsid.
  do {
    nsid = get_nsid();
    hash = get_nsid_hash(ns);
    found = GET_NS_BY_HASH(srv, hash) != NULL;
  } while (found);
  // Initialize and register ns.
  elton_rpc_ns_init(ns, s, nsid, true);
  ADD_NS(ns);
  return 0;
}

static void rpc_close_nowait(struct elton_rpc_server *srv) {
  // todo: impl
  // todo: 既存のセッションは強制切断する?
}

static int rpc_close(struct elton_rpc_server *srv) {
  rpc_close_nowait(srv);
  // todo: Wait for close.
}

static struct elton_rpc_operations rpc_ops = {
    .listen = rpc_listen,
    .start_umh = rpc_start_umh,
    .new_session = rpc_new_session,
    .close_nowait = rpc_close_nowait,
    .close = rpc_close,
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

// Send a packet.  MUST acquire sock_write_lock and ns->lock before call it.
static int ns_send_packet_without_lock(struct elton_rpc_ns *ns, int flags,
                                       int struct_id, void *data) {
  int error = 0;
  struct raw_packet *raw;
  struct packet pkt = {
      .struct_id = struct_id,
      .data = data,
  };
  struct xdr_encoder enc;
  char buff[25];
  loff_t offset;
  ssize_t n;

  GOTO_IF(error, elton_rpc_encode_packet(&pkt, &raw));
  raw->session_id = ns->nsid;
  raw->flags = flags;

  GOTO_IF(error_encode, default_encoder_init(&enc, buff, sizeof(buff)));
  enc.enc_op->u64(&enc, raw->size);
  enc.enc_op->u64(&enc, raw->session_id);
  enc.enc_op->u8(&enc, raw->flags);
  enc.enc_op->u64(&enc, raw->struct_id);
  if (enc.error) {
    error = enc.error;
    NS_ERR(ns, "failed to encode with error code %d", error);
    goto error_encode;
  }

  offset = 0;
  while (offset < sizeof(buff)) {
    n = WRITE_SOCK(ns->session->sock, buff, sizeof(buff), &offset);
    if (n < 0) {
      error = n;
      NS_ERR(ns, "write error %d", error);
      goto error_write;
    }
  }
  NS_DEBUG(ns, "sent a struct: struct_id=%d, flags=%d", struct_id, flags);

error_write:
error_encode:
  raw->free(raw);
error:
  return error;
}

static int ns_send_struct(struct elton_rpc_ns *ns, int struct_id, void *data) {
  int error = 0;
  int flags = 0;

  mutex_lock(&ns->session->sock_write_lock);
  spin_lock(&ns->lock);
  if (!ns->sendable) {
    // Can not send data because it is closed.
    error = -ELTON_RPC_ALREADY_CLOSED;
    goto error_precondition;
  }
  if (!ns->established)
    flags |= ELTON_SESSION_FLAG_CREATE;
  if (struct_id == ELTON_RPC_ERROR_ID)
    flags |= ELTON_SESSION_FLAG_ERROR;
  spin_unlock(&ns->lock);

  error = ns_send_packet_without_lock(ns, flags, struct_id, data);

  spin_lock(&ns->lock);
  if (!error)
    ns->established = true;
error_precondition:
  spin_unlock(&ns->lock);
  mutex_unlock(&ns->session->sock_write_lock);
  return error;
}

static int ns_send_error(struct elton_rpc_ns *ns,
                         struct elton_rpc_error *error) {
  return ns_send_struct(ns, ELTON_RPC_ERROR_ID, error);
}

static int ns_recv_struct(struct elton_rpc_ns *ns, u64 struct_id, void **data) {
  int error = 0;
  struct raw_packet *raw = NULL;

  NS_DEBUG(ns, "waiting a struct to be received");
  elton_rpc_dequeue(&ns->q, &raw);

  if (raw->flags & ELTON_SESSION_FLAG_CLOSE) {
    // todo: race conditionの可能性あり
    // dequeueしてからflagsをチェックする間は、排他制御が行われていない.
    spin_lock(&ns->lock);
    ns->receivable = false;
    spin_unlock(&ns->lock);
    NS_DEBUG(ns, "receivable=false");
  }

  if (raw->struct_id != struct_id) {
    // Unexpected struct.
    error = -ELTON_RPC_DIFF_TYPE;
    NS_ERR(ns, "unexpected struct is received: expected=%llu, actual=%llu",
           struct_id, raw->struct_id);
    goto error_dequeue;
  }

  GOTO_IF(error_dequeue, elton_rpc_decode_packet(raw, data));
  NS_DEBUG(ns, "received a struct");

error_dequeue:
  if (raw)
    raw->free(raw);
  return error;
}

static int ns_close(struct elton_rpc_ns *ns) {
  int error = 0;
  struct elton_rpc_ping ping;

  mutex_lock(&ns->session->sock_write_lock);
  spin_lock(&ns->lock);
  if (!ns->established)
    // This session is not established. Does not need to send a close packet.
    goto out;
  if (!ns->sendable) {
    // Already closed?
    error = -ELTON_RPC_ALREADY_CLOSED;
    goto out;
  }
  spin_unlock(&ns->lock);

  error = ns_send_packet_without_lock(ns, ELTON_RPC_PACKET_FLAG_CLOSE,
                                      ELTON_RPC_PING_ID, &ping);

  spin_lock(&ns->lock);
  ns->sendable = false;
out:
  spin_unlock(&ns->lock);
  mutex_unlock(&ns->session->sock_write_lock);
  return error;
}

static bool ns_is_sendable(struct elton_rpc_ns *ns) {
  bool sendable;
  spin_lock(&ns->lock);
  sendable = ns->sendable;
  spin_unlock(&ns->lock);
  return sendable;
}

static bool ns_is_receivable(struct elton_rpc_ns *ns) {
  bool receivable;
  spin_lock(&ns->lock);
  receivable = ns->receivable;
  spin_unlock(&ns->lock);
  return receivable;
}

static struct elton_rpc_ns_operations ns_op = {
    .send_struct = ns_send_struct,
    .send_error = ns_send_error,
    .recv_struct = ns_recv_struct,
    .close = ns_close,
    .is_sendable = ns_is_sendable,
    .is_receivable = ns_is_receivable,
};

static void elton_rpc_ns_init(struct elton_rpc_ns *ns,
                              struct elton_rpc_session *s, u64 nsid,
                              bool is_client) {
  ns->session = s;
  ns->nsid = nsid;
  elton_rpc_queue_init(&ns->q);
  spin_lock_init(&ns->lock);
  if (is_client) {
    // client
    ns->established = false;
    ns->sendable = true;
    ns->receivable = false;
  } else {
    // server
    ns->established = true;
    ns->sendable = true;
    ns->receivable = true;
  }
  ns->ops = &ns_op;
}
