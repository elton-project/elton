#ifndef _ELTON_RPC_SERVER_H
#define _ELTON_RPC_SERVER_H

#include <elton/rpc/packet.h>
#include <elton/rpc/queue.h>
#include <linux/hashtable.h>
#include <linux/types.h>

#define ELTON_UMH_SOCK "/run/elton.sock"
#define ELTON_SESSIONS_HASH_SIZE 3
#define ELTON_NS_HASH_SIZE 12

struct elton_rpc_server {
  // Path to UNIX domain socket.
  const char *socket_path;
  // Server socket.  Use it for listening only.
  struct socket *sock;
  // For main worker thread.
  // It accepts new connection and start session worker.
  struct task_struct *task;
  struct mutex task_lock;
  // Hash table for sessions.  Please do not put many sessions to prevent
  // performance degradation.  We supposes that number of sessions is 1 to 3.
  //
  // Key: Session ID.
  // Value: struct elton_rpc_session
  DECLARE_HASHTABLE(ss_table, ELTON_SESSIONS_HASH_SIZE);
  // Hash table for nested sessions.  We supposes that number of nested sessions
  // is 1 to 1024.
  //
  // Key: Nested Session ID
  // Value: struct elton_rpc_ns
  DECLARE_HASHTABLE(nss_table, ELTON_NS_HASH_SIZE);

  struct elton_rpc_operations *ops;
};
struct elton_rpc_session {
  struct elton_rpc_server *server;
  // Required for building a hashtable.
  struct hlist_node _hash;
  // Session ID.
  u8 sid;
  // Socket for the RPC session.  It can only be sent/received packets.
  struct socket *sock;
  struct mutex sock_write_lock;

  // For session worker thread.
  // It reads everything from the socket and decode to raw_packets.
  struct task_struct *task;
  struct mutex task_lock;
  // Queue for received packets.
  // The nested session reads the packet from it.
  struct elton_rpc_queue q;
};
struct elton_rpc_ns {
  struct elton_rpc_session *session;
  // Required for building a hashtable.
  struct hlist_node _hash;
  // Nested Session ID
  u64 nsid;
  struct elton_rpc_ns_operations *ops;
};

struct elton_rpc_operations {
  // Start server.
  int (*listen)(struct elton_rpc_server *);
  // Start UMH (User Mode Helper) process.
  int (*start_umh)(struct elton_rpc_server *);
  // Create new nested session.
  int (*new_session)(struct elton_rpc_server *, struct elton_rpc_ns *);
  // Request to shutdown the server.
  void (*close_nowait)(struct elton_rpc_server *);
  // Shutdown server and wait it.
  int (*close)(struct elton_rpc_server *);
};
struct elton_rpc_ns_operations {
  int (*send_struct)(struct elton_rpc_ns *ns, int struct_id, void *data);
  int (*send_error)(struct elton_rpc_ns *ns, int error);
  int (*recv_struct)(struct elton_rpc_ns *ns, int struct_id, void *data);
  int (*close)(struct elton_rpc_ns *ns);
  bool (*is_sendable)(struct elton_rpc_ns *ns);
  bool (*is_receivable)(struct elton_rpc_ns *ns);
};

int elton_rpc_server_init(struct elton_rpc_server *server, char *socket_path);

#endif // _ELTON_RPC_SERVER_H
