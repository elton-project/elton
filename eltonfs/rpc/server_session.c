#define ELTON_LOG_PREFIX "[rpc/server sid=%d] "
#define ELTON_LOG_PREFIX_ARGS , (int)(s ? s->sid : -1)

#include <elton/rpc/_server.h>

static struct elton_rpc_setup2 setup2 = {
    .error = 0,
    .reason = "",
    .server_name = "eltonfs",
    .version_major = 0,    // todo
    .version_minor = 0,    // todo
    .version_revision = 0, // todo
};

void elton_rpc_session_init(struct elton_rpc_session *s,
                            struct elton_rpc_server *server, u8 sid,
                            struct socket *sock) {
  s->server = server;
  s->sid = sid;
  s->sock = sock;
  mutex_init(&s->sock_write_lock);
  s->sock_wb = NULL;
  s->sock_wb_size = 0;
  s->task = NULL;
  mutex_init(&s->task_lock);
}

// Receive setup1 packet from client and validate it.
static int rpc_session_setup1(struct elton_rpc_session *s,
                              struct elton_rpc_setup1 **setup1) {
  int error;
  DEBUG("waiting setup1 ...");
  RETURN_IF(
      rpc_sock_read_packet(s->sock, ELTON_RPC_SETUP1_ID, (void **)setup1));
  DEBUG("received setup1 from client");

  DEBUG("validating setup1");
  // todo: check setup1.
  return 0;
}
// Send setup1 packet to client.
static int rpc_session_setup2(struct elton_rpc_session *s) {
  int error = 0;
  struct packet pk = {
      .struct_id = ELTON_RPC_SETUP2_ID,
      .data = &setup2,
  };
  struct raw_packet *raw = NULL;

  DEBUG("preparing setup2 ...");
  GOTO_IF(error_setup2, elton_rpc_encode_packet(&pk, &raw, 0, 0));
  BUG_ON(raw == NULL);
  BUG_ON(raw->data == NULL);

  DEBUG("sending setup2 ...");
  GOTO_IF(error_setup2, rpc_sock_write_raw_packet(s->sock, raw));
  DEBUG("sent setup2");

error_setup2:
  if (raw)
    raw->free(raw);
  return error;
}

static void free_ns(struct elton_rpc_ns *ns) { kfree(ns); }
static int alloc_ns(struct elton_rpc_ns **ns, struct elton_rpc_session *s,
                    u64 nsid, bool is_client) {
  int error;
  struct elton_rpc_ns *out;

  out = (struct elton_rpc_ns *)kmalloc(sizeof(struct elton_rpc_ns), GFP_KERNEL);
  if (out == NULL) {
    ERR("failed to allocate elton_rpc_ns object");
    RETURN_IF(-ENOMEM);
  }
  elton_rpc_ns_init(out, s, nsid, false, free_ns);
  *ns = out;
  return 0;
}

// Qneueue raw packet to ns->queue.
static int rpc_session_enqueue_raw_packet(struct elton_rpc_session *s,
                                          struct raw_packet *raw) {
  int error = 0;
  struct elton_rpc_ns *ns = NULL;
  BUG_ON(raw == NULL);

  spin_lock(&s->server->nss_table_lock);
  ns = GET_NS_NOLOCK(s, raw->session_id);
  if (ns) {
    // Enqueue it.
    GOTO_IF(out_unlock, elton_rpc_enqueue(&ns->q, raw));
  } else if (raw->flags & ELTON_SESSION_FLAG_CREATE) {
    struct new_ns_handler_args *handler_args;
    struct task_struct *handler_task;

    // Create session.
    GOTO_IF(out_unlock, alloc_ns(&ns, s, raw->session_id, false));

    // Start handler.
    GOTO_IF(out_unlock,
            new_ns_handler_args(&handler_args, ns, raw->struct_id, raw->flags));
    handler_task = kthread_create(elton_rpc_new_ns_handler, handler_args,
                                  "elton-rpc [%llu]", ns->nsid);
    if (IS_ERR(handler_task))
      GOTO_IF(out_unlock, PTR_ERR(handler_task));
    ns->handler_task = handler_task;

    DEBUG("created new session by umh");
    ADD_NS_NOLOCK(ns);
    wake_up_process(ns->handler_task);

    // Enqueue it.
    GOTO_IF(out_unlock, elton_rpc_enqueue(&ns->q, raw));

    ns = NULL;
  } else {
    ERR("ns not found: s=%px, ns=%px, raw=%px, nsid=%llu, flags=%hhu, "
        "struct_id=%llu",
        s, ns, raw, raw->session_id, raw->flags, raw->struct_id);
    BUG();
    // Unreachable
    return -ENOTRECOVERABLE;
  }

out_unlock:
  spin_unlock(&s->server->nss_table_lock);
  return error;
}

int rpc_session_pinger(void *_s) {
  int error = 0;
  struct elton_rpc_session *s = (struct elton_rpc_session *)_s;
  struct elton_rpc_ping ping = {};
  struct elton_rpc_ns ns;
  struct elton_rpc_ping *recved_ping;

  while (true) {
    // nsはスタック領域に確保しているので、メモリ解放は不要。
    GOTO_IF(error_ns, s->server->ops->new_session(s->server, &ns, NULL));
    GOTO_IF(error_send, ns.ops->send_struct(&ns, ELTON_RPC_PING_ID, &ping));
    GOTO_IF(error_recv,
            ns.ops->recv_struct(&ns, ELTON_RPC_PING_ID, (void **)&recved_ping));
    GOTO_IF(error_close, ns.ops->close(&ns));
    DEBUG("ping OK");
    msleep_interruptible(1000);
  }

error_recv:
error_send:
  RETURN_IF(ns.ops->close(&ns));
error_close:
error_ns:
  return error;
}

bool validate_received_packet(struct raw_packet *raw) {
  if (!(ELTON_RPC_MIN_NSID <= raw->session_id &&
        raw->session_id <= ELTON_RPC_MAX_NSID)) {
    return false;
  }
  if (raw->flags & ELTON_SESSION_FLAG_CREATE &&
      !(ELTON_RPC_MIN_CLIENT_NSID <= raw->session_id &&
        raw->session_id <= ELTON_RPC_MAX_CLIENT_NSID)) {
    return false;
  }
  return true;
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
int rpc_session_worker(void *_s) {
  int error = 0;
  struct elton_rpc_session *s = (struct elton_rpc_session *)_s;
  struct elton_rpc_setup1 *setup1 = NULL;
  struct task_struct *pinger = NULL;

  GOTO_IF(error_setup1, rpc_session_setup1(s, &setup1));
  GOTO_IF(error_setup2, rpc_session_setup2(s));

  INFO("established connection (client=%s)",
       setup1->client_name ? setup1->client_name : "no-name client");

  // Start health check worker.
  pinger = (struct task_struct *)kthread_run(rpc_session_pinger, s,
                                             "elton-ping [%d]", s->sid);

  // Receive data from client until socket is closed.
  for (;;) {
    struct raw_packet *raw = NULL;

    GOTO_IF(error_recv, rpc_sock_read_raw_packet(s->sock, &raw));
    if (!validate_received_packet(raw)) {
      ERR("invalid nsid: s=%px, raw=%px, nsid=%llu", s, raw, raw->session_id);
      BUG();
    }

    DEBUG("received a packet: struct_id=%llu, flags=%hhu, size=%zu",
          raw->struct_id, raw->flags, raw->size);
    GOTO_IF(error_enqueue, rpc_session_enqueue_raw_packet(s, raw));
    continue;

  error_enqueue:
    if (raw)
      raw->free(raw);
    raw = NULL;
    goto error_recv;
  }

error_recv:
  // TODO: Notify pinger that the session is going to stop.
  // TODO: Wait for pinger thread.
error_setup2:
  if (setup1)
    elton_rpc_free_decoded_data(setup1);
error_setup1:
  INFO("stopping session worker");
  kfree(s->sock);
  s->sock = NULL;
  // Unregister from s->server->ss_table.
  DELETE_SESSION(s);
  return error;
}
