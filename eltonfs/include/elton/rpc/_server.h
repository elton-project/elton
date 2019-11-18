// Define internal functions and macros

#include <elton/assert.h>
#include <elton/error.h>
#include <elton/rpc/packet.h>
#include <elton/rpc/server.h>
#include <elton/rpc/struct.h>
#include <elton/xdr/interface.h>
#include <linux/kthread.h>
#include <linux/types.h>
#include <linux/un.h>
#include <net/sock.h>

#define LISTEN_LENGTH 4
#define SOCK_PROTO_NAME(sock) (sock)->sk->sk_prot_creator->name

void elton_rpc_session_init(struct elton_rpc_session *s,
                            struct elton_rpc_server *server, u8 sid,
                            struct socket *sock);
int rpc_session_worker(void *_s);
void elton_rpc_ns_init(struct elton_rpc_ns *ns, struct elton_rpc_session *s,
                       u64 nsid, bool is_client);
u64 get_nsid(void);
u64 get_nsid_hash(struct elton_rpc_ns *ns);
u64 get_nsid_hash_by_values(u8 session_id, u64 nsid);

// Handler for nested session created by UMH.
int elton_rpc_new_ns_handler(void *_args);
// Argument for elton_rpc_new_ns_handler().
struct new_ns_handler_args {
  struct elton_rpc_ns *ns;
  u64 struct_id;
  u8 flags;
  void (*free)(struct new_ns_handler_args *args);
};
int new_ns_handler_args(struct new_ns_handler_args **args,
                        struct elton_rpc_ns *ns, u64 struct_id, u8 flags);

int rpc_sock_read_packet(struct socket *sock, u64 struct_id, void **out);
int rpc_sock_read_raw_packet(struct socket *sock, struct raw_packet **out);
int rpc_sock_write_raw_packet(struct socket *sock, struct raw_packet *raw);

// Get an entry from hashtable.  If entry is not found, returns NULL.
//
// Arguments:
//   hashtable:
//   type: the type name of entry.
//   member: the name of the hlist_node.
//   key_member: the field name that the hash key is stored.
//   key: actual key.
#define HASH_GET(hashtable, type, member, key_member, key)                     \
  ({                                                                           \
    type *obj = NULL;                                                          \
    hash_for_each_possible((hashtable), obj, member, (key)) {                  \
      if (obj->key_member == (key))                                            \
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
// MUST acquire the server->nss_table_lock before call it.
#define ADD_NS_NOLOCK(ns)                                                      \
  do {                                                                         \
    u64 hash = get_nsid_hash(ns);                                              \
    hash_add((ns)->session->server->nss_table, &(ns)->_hash, hash);            \
  } while (0)
// Add nested session to server.
// MUST NOT call when acquired the server->nss_table_lock.
#define ADD_NS(ns)                                                             \
  do {                                                                         \
    spin_lock(&(ns)->session->server->nss_table_lock);                         \
    ADD_NS_NOLOCK(ns);                                                         \
    spin_unlock(&(ns)->session->server->nss_table_lock);                       \
  } while (0)
// Get elton_rpc_ns by hash value of nsid.
// MUST acquire the server->nss_table_lock before call it.
static inline struct elton_rpc_ns *GET_NS_NOLOCK(struct elton_rpc_session *s,
                                                 u64 nsid) {
  struct elton_rpc_ns *ns = NULL;
  u64 hash = get_nsid_hash_by_values(s->sid, nsid);
  bool found = false;
  hash_for_each_possible(s->server->nss_table, ns, _hash, hash) {
    if (ns->nsid == nsid && ns->session == s) {
      found = true;
      break;
    }
  }
  if (!found)
    ns = NULL;
  return ns;
}
// Get elton_rpc_ns by hash value of nsid.
// MUST NOT call when acquired the server->nss_table_lock.
static inline struct elton_rpc_ns *GET_NS(struct elton_rpc_session *s,
                                          u64 nsid) {
  struct elton_rpc_ns *ns;
  spin_lock(&s->server->nss_table_lock);
  ns = GET_NS_NOLOCK(s, nsid);
  spin_unlock(&s->server->nss_table_lock);
  return ns;
}
