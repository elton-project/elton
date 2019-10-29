#ifndef _ELTON_RPC_SERVER_H
#define _ELTON_RPC_SERVER_H

#include <elton/rpc/packet.h>
#include <linux/types.h>

#define ELTON_UMH_SOCK "/run/elton.sock"

struct elton_rpc_server {
  // Path to UNIX domain socket.
  const char *socket_path;
  // TODO
  struct elton_rpc_operations *ops;
};
struct elton_rpc_session {
  struct elton_rpc_server *server;
  // Nested Session ID
  u64 nsid;
  // TODO
  struct elton_rpc_session_operations *ops;
};

struct elton_rpc_operations {
  // Start server.
  int (*listen)(struct elton_rpc_server *);
  // Start UMH (User Mode Helper) process.
  int (*start_umh)(struct elton_rpc_server *);
  // Create new nested session.
  int (*new_session)(struct elton_rpc_server *, struct elton_rpc_session *);
  // Request to shutdown the server.
  void (*close_nowait)(struct elton_rpc_server *);
  // Shutdown server and wait it.
  int (*close)(struct elton_rpc_server *);
};
struct elton_rpc_session_operations {
  int (*send_struct)(struct elton_rpc_session *, int struct_id, void *data);
  int (*send_error)(struct elton_rpc_session *, int error);
  int (*recv_struct)(struct elton_rpc_session *, int struct_id, void *data);
  int (*close)(struct elton_rpc_session *);
  bool (*is_sendable)(struct elton_rpc_session *);
  bool (*is_receivable)(struct elton_rpc_session *);
};

int elton_rpc_server_init(struct elton_rpc_server *server, char *socket_path);

#endif // _ELTON_RPC_SERVER_H
