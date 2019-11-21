#define ELTON_LOG_PREFIX "[rpc/server sid=%d nsid=%llu] "
#define ELTON_LOG_PREFIX_ARGS , (int)(ns ? ns->session->sid : -1), ns->nsid

#include <elton/rpc/_server.h>

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
u64 get_nsid(void) {
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

u64 get_nsid_hash_by_values(u8 session_id, u64 nsid) {
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
u64 get_nsid_hash(struct elton_rpc_ns *ns) {
  return get_nsid_hash_by_values(ns->session->sid, ns->nsid);
}

bool validate_send_packet(struct raw_packet *raw) {
  if (!(ELTON_RPC_MIN_NSID <= raw->session_id &&
        raw->session_id <= ELTON_RPC_MAX_NSID)) {
    return false;
  }
  if (raw->flags & ELTON_SESSION_FLAG_CREATE &&
      !(ELTON_RPC_MIN_SERVER_NSID <= raw->session_id &&
        raw->session_id <= ELTON_RPC_MAX_SERVER_NSID)) {
    return false;
  }
  return true;
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

  GOTO_IF(error, elton_rpc_encode_packet(&pkt, &raw, ns->nsid, flags));
  GOTO_IF(error_write, rpc_sock_write_raw_packet(ns->session->sock, raw));
  DEBUG("sent a struct: struct_id=%d, flags=%d", struct_id, flags);

error_write:
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
    GOTO_IF(error_precondition, -ELTON_RPC_ALREADY_CLOSED);
  }
  if (!ns->established)
    flags |= ELTON_SESSION_FLAG_CREATE;
  if (struct_id == ELTON_RPC_ERROR_ID)
    flags |= ELTON_SESSION_FLAG_ERROR;
  spin_unlock(&ns->lock);

  error = ns_send_packet_without_lock(ns, flags, struct_id, data);

  spin_lock(&ns->lock);
  if (!error) {
    ns->established = true;
    ns->receivable = true;
  }
error_precondition:
  spin_unlock(&ns->lock);
  mutex_unlock(&ns->session->sock_write_lock);

  RETURN_IF(error);
  return 0;
}

static int ns_send_error(struct elton_rpc_ns *ns, struct elton_rpc_error *e) {
  int error;
  RETURN_IF(ns_send_struct(ns, ELTON_RPC_ERROR_ID, e));
  return 0;
}

static int ns_recv_struct(struct elton_rpc_ns *ns, u64 struct_id, void **data) {
  int error = 0;
  struct raw_packet *raw = NULL;

  DEBUG("waiting a struct to be received");
  elton_rpc_dequeue(&ns->q, &raw);

  if (raw->flags & ELTON_SESSION_FLAG_CLOSE) {
    // todo: race conditionの可能性あり
    // dequeueしてからflagsをチェックする間は、排他制御が行われていない.
    spin_lock(&ns->lock);
    ns->receivable = false;
    spin_unlock(&ns->lock);
    DEBUG("receivable=false");
  }

  if (raw->struct_id != struct_id) {
    // Unexpected struct.
    ERR("unexpected struct is received: expected=%llu, actual=%llu", struct_id,
        raw->struct_id);
    GOTO_IF(error_dequeue, -ELTON_RPC_DIFF_TYPE);
  }

  GOTO_IF(error_dequeue, elton_rpc_decode_packet(raw, data));
  DEBUG("received a struct");

error_dequeue:
  if (raw)
    raw->free(raw);
  return error;
}

static int ns_close(struct elton_rpc_ns *ns) {
  int error = 0;
  struct elton_rpc_ping ping;
  bool receivable = false;
  struct elton_rpc_session *s = ns->session;

  mutex_lock(&s->sock_write_lock);
  spin_lock(&ns->lock);
  if (!ns->established)
    // This session is not established. Does not need to send a close packet.
    goto out;
  if (!ns->sendable)
    // Already closed?
    GOTO_IF(out, -ELTON_RPC_ALREADY_CLOSED);
  ns->sendable = false;

  if (!ns->receivable) {
    // Closed from both direction.
    //
    // 1. Closed from UMH.  Send a packet with close flags.
    // 2. Closed from kmod.  Send a packet with close flags and release memory.
    //      ↓↓↓
    GOTO_IF(out, ns_send_packet_without_lock(ns, ELTON_RPC_PACKET_FLAG_CLOSE,
                                             ELTON_RPC_PING_ID, &ping));
    spin_unlock(&ns->lock);
    DELETE_NS(ns);
    // 3. UMH receives a packet with close flags.  Should release memory.
    goto out_without_ns_unlock;
  } else {
    // Closed from kmod.  Wait for close ns from UMH.
    //
    // 1. Closed from kmod.  Send a packet with close flags.
    // 2. Closed from UMH.  Send a packet with close flags and release memory.
    // 3. Receive a packet with close flags from UMH.  Should release memory.
    //      ↓↓↓
    spin_unlock(&ns->lock);
    // Wait for receive a packet with close flags.
    do {
      bool closed, is_ping;
      struct raw_packet *raw = NULL;

      CHECK_ERROR(elton_rpc_dequeue(&ns->q, &raw));
      BUG_ON(raw == NULL);

      closed = raw->flags & ELTON_SESSION_FLAG_CLOSE;
      is_ping = raw->struct_id == ELTON_RPC_PING_ID;

      WARN(!closed,
           "receive a packet after closed: ns=%px, nsid=%llu, struct_id=%llu",
           ns, ns->nsid, raw->struct_id);
      WARN(closed && is_ping,
           "nested session closed with unexpected struct type: ns=%px, "
           "nsid=%llu, struct_id=%llu",
           ns, ns->nsid, raw->struct_id);

      raw->free(raw);
      if (closed)
        break;
    } while (1);

    // Received a closed packet.  Should release memory.
    DELETE_NS(ns);
    goto out_without_ns_unlock;
  }

  // Unreachable.
  BUG();

out:
  spin_unlock(&ns->lock);
out_without_ns_unlock:
  mutex_unlock(&s->sock_write_lock);
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

void elton_rpc_ns_init(struct elton_rpc_ns *ns, struct elton_rpc_session *s,
                       u64 nsid, bool is_client,
                       void (*free)(struct elton_rpc_ns *)) {
  ns->session = s;
  ns->nsid = nsid;
  elton_rpc_queue_init(&ns->q);
  ns->handler_task = NULL;
  ns->free = free;
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
