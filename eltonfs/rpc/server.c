#define ELTON_LOG_PREFIX "[rpc/server] "

#include <elton/rpc/_server.h>

static int rpc_sock_set_file(struct socket *newsock, char *name) {
  int error;
  int newfd;
  struct file *newfile;

  newfd = get_unused_fd_flags(0);
  if (newfd < 0)
    GOTO_IF(error_newfd, newfd);

  newfile = sock_alloc_file(newsock, 0, name);
  if (IS_ERR(newfile))
    GOTO_IF(error_alloc_file, PTR_ERR(newfile));

  fd_install(newfd, newfile);
  return 0;

error_alloc_file:
  put_unused_fd(newfd);
error_newfd:
  return error;
}

static int rpc_sock_accpet(struct socket *sock, struct socket **new) {
  int error;
  struct socket *newsock = NULL;
  GOTO_IF(error, kernel_accept(sock, &newsock, sock->file->f_flags));
  GOTO_IF(error_set_file, rpc_sock_set_file(newsock, SOCK_PROTO_NAME(sock)));
  *new = newsock;
  return 0;

error_set_file:
  BUG_ON(newsock == NULL);
  sock_release(newsock);
error:
  return error;
}

// Listen and accept new connection.
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
  struct sockaddr_un addr;

  // Initialize socket and listen.
  GOTO_IF(error, sock_create(AF_UNIX, SOCK_STREAM, 0, &srv->sock));
  addr.sun_family = AF_UNIX;
  strncpy(addr.sun_path, srv->socket_path, UNIX_PATH_MAX);
  GOTO_IF(error_sock, rpc_sock_set_file(srv->sock, SOCK_PROTO_NAME(srv->sock)));
  GOTO_IF(error_sock, srv->sock->ops->bind(srv->sock, (struct sockaddr *)&addr,
                                           sizeof(struct sockaddr_un)));
  // TODO: Change socket file mode. We can safely change the mode of socket file
  // at this time because it is not listened yet.
  GOTO_IF(error_sock, srv->sock->ops->listen(srv->sock, LISTEN_LENGTH));
  DEBUG("listened UNIX Domain Socket");

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
      ERR("master: failed to allocate elton_rpc_session");
      GOTO_IF(error_accept, -ENOMEM);
    }
    elton_rpc_session_init(s, srv, session_id, newsock);

    // Start session worker.
    task = (struct task_struct *)kthread_run(rpc_session_worker, s,
                                             "elton-rpc [%d]", session_id);
    if (IS_ERR(task)) {
      ERR("master: kthread_run returns an error %d", error);
      GOTO_IF(error_kthread, PTR_ERR(task));
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

error_sock:
  sock_release(srv->sock);
error:
  ERR("master: stopped");
  return error;
}

// Start a worker that serve RPC in the background.
static int rpc_start_worker(struct elton_rpc_server *s) {
  int error = 0;
  struct task_struct *task;

  // Start master worker.
  task = kthread_run(rpc_master_worker, s, "elton-rpc [master]");
  if (IS_ERR(task)) {
    ERR("kthread_run returns an error %d", error);
    GOTO_IF(error, PTR_ERR(task));
  }
  INFO("started master worker");
  mutex_lock(&s->task_lock);
  s->task = task;
  mutex_unlock(&s->task_lock);
  return 0;

error:
  return error;
}

// Start UMH (User Mode Helper) in the background.
static int rpc_start_umh(struct elton_rpc_server *s) {
  int error = 0;
  struct subprocess_info *info;
  char *argv[] = {
      // Redirect stdin/stdout/stderr by shell.
      "/bin/sh",
      "-c",
      ELTONFS_HELPER " --socket " ELTONFS_HELPER_SOCK " 2>&1 |"
                     "logger -t " ELTONFS_HELPER_LOG_TAG,
      NULL,
  };
  char *envp[] = {
      "HOME=/",
      "TERM=linux",
      "PATH=" PATH_ENV,
      NULL,
  };

  info = call_usermodehelper_setup(argv[0], argv, envp, GFP_KERNEL, NULL, NULL,
                                   NULL);
  if (info == NULL)
    RETURN_IF(-ENOMEM);

  // Register subprocess_info to server.
  mutex_lock(&s->umh_info_lock);
  s->umh_info = info;
  mutex_unlock(&s->umh_info_lock);

  error = call_usermodehelper_exec(info, UMH_WAIT_EXEC);
  if (error) {
    ERR("failed to start UMH with error %d", error);
    RETURN_IF(error);
  }
  INFO("start " ELTONFS_HELPER);
  return 0;
}

static int rpc_new_session(struct elton_rpc_server *srv,
                           struct elton_rpc_ns *ns,
                           void (*free)(struct elton_rpc_ns *)) {
  int error;
  struct elton_rpc_session *s = NULL;
  u64 nsid;
  bool found;

  // Select a session.
  FOR_EACH_SESSIONS(srv, s, break);
  if (s == NULL)
    // Session not found.  Elton needs at least one session to create nested
    // session.
    RETURN_IF(-EINVAL);

  // Find the unused nsid.
  do {
    nsid = get_nsid();
    found = GET_NS(s, nsid) != NULL;
  } while (found);
  // Initialize and register ns.
  elton_rpc_ns_init(ns, s, nsid, true, free);
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
    .start_worker = rpc_start_worker,
    .start_umh = rpc_start_umh,
    .new_session = rpc_new_session,
    .close_nowait = rpc_close_nowait,
    .close = rpc_close,
};

int elton_rpc_server_init(struct elton_rpc_server *server, char *socket_path) {
  if (socket_path == NULL)
    socket_path = ELTONFS_HELPER_SOCK;
  server->socket_path = socket_path;
  server->sock = NULL;
  server->task = NULL;
  mutex_init(&server->task_lock);
  server->umh_info = NULL;
  mutex_init(&server->umh_info_lock);
  hash_init(server->ss_table);
  spin_lock_init(&server->ss_table_lock);
  hash_init(server->nss_table);
  spin_lock_init(&server->nss_table_lock);
  server->ops = &rpc_ops;
  return 0;
}
