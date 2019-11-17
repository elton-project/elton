#ifndef _ELTON_RPC_SERVER_H
#define _ELTON_RPC_SERVER_H

#include <elton/rpc/packet.h>
#include <elton/rpc/queue.h>
#include <elton/rpc/struct.h>
#include <linux/hashtable.h>
#include <linux/types.h>

#define ELTON_UMH_SOCK "/run/elton.sock"
#define ELTON_SESSIONS_HASH_SIZE 3
#define ELTON_NS_HASH_SIZE 12

#define ELTON_RPC_MIN_NSID 1UL
#define ELTON_RPC_MAX_NSID ((1UL << 32) - 1UL)
#define ELTON_RPC_MIN_SERVER_NSID 1UL
#define ELTON_RPC_MAX_SERVER_NSID ((1UL << 31) - 1UL)
#define ELTON_RPC_MIN_CLIENT_NSID (1UL << 31)
#define ELTON_RPC_MAX_CLIENT_NSID ((1UL << 32) - 1UL)

#define ELTON_SESSION_FLAG_CREATE 1
#define ELTON_SESSION_FLAG_CLOSE 2
#define ELTON_SESSION_FLAG_ERROR 3

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
  struct spinlock ss_table_lock;
  // Hash table for nested sessions.  We supposes that number of nested sessions
  // is 1 to 1024.
  //
  // Key: Nested Session ID
  // Value: struct elton_rpc_ns
  DECLARE_HASHTABLE(nss_table, ELTON_NS_HASH_SIZE);
  struct spinlock nss_table_lock;

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
  // Buffer for writing to socket.  MUST acquire sock_write_lock before use it.
  char *sock_wb;
  size_t sock_wb_size;

  // For session worker thread.
  // It reads everything from the socket and decode to raw_packets.
  struct task_struct *task;
  struct mutex task_lock;
};
struct elton_rpc_ns {
  struct elton_rpc_session *session;
  // Required for building a hashtable.
  struct hlist_node _hash;
  // Nested Session ID
  u64 nsid;
  // Queue for received packets.
  // The nested session reads the packet from it.
  struct elton_rpc_queue q;

  // MUST acquire a lock before accessing to these fields.
  struct spinlock lock;
  bool established;
  bool sendable;
  bool receivable;

  struct elton_rpc_ns_operations *ops;
};

struct elton_rpc_operations {
  // Start RPC server.
  int (*start_worker)(struct elton_rpc_server *);
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
  int (*send_error)(struct elton_rpc_ns *ns, struct elton_rpc_error *err);
  int (*recv_struct)(struct elton_rpc_ns *ns, u64 struct_id, void **data);
  int (*close)(struct elton_rpc_ns *ns);
  bool (*is_sendable)(struct elton_rpc_ns *ns);
  bool (*is_receivable)(struct elton_rpc_ns *ns);
};

int elton_rpc_server_init(struct elton_rpc_server *server, char *socket_path);

#endif // _ELTON_RPC_SERVER_H
