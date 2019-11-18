#define ELTON_LOG_PREFIX "[rpc/server sid=%d nsid=%llu] "
#define ELTON_LOG_PREFIX_ARGS , (int)(ns ? ns->session->sid : -1), ns->nsid
#include <elton/rpc/_server.h>

static int ping_handler(struct elton_rpc_ns *ns) {
  int error;
  struct elton_rpc_ping *recved_ping;
  struct elton_rpc_ping ping = {};

  DEBUG("received ping from UMH.");
  GOTO_IF(error_recv,
          ns->ops->recv_struct(ns, ELTON_RPC_PING_ID, (void **)&recved_ping));
  GOTO_IF(error_send, ns->ops->send_struct(ns, ELTON_RPC_PING_ID, &ping));

error_send:
error_recv:
  RETURN_IF(ns->ops->close(ns));
  return error;
}

int elton_rpc_new_ns_handler(void *_args) {
  int error;
  struct new_ns_handler_args *args = (struct new_ns_handler_args *)_args;
  struct elton_rpc_ns *ns = args->ns;

  switch (args->struct_id) {
  case ELTON_RPC_PING_ID:
    RETURN_IF(ping_handler(ns));
    break;
  default:
    ERR("not supported type: args=%px, sid=%llu", args, args->struct_id);
    BUG();
    // Unreachable.
  }
}
