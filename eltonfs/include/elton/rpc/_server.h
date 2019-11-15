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
// MUST acquire the server->nss_table_lock before call it.
#define ADD_NS_NOLOCK(ns)                                                      \
  hash_add((ns)->session->server->nss_table, &(ns)->_hash, (ns)->nsid)
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
#define GET_NS_BY_HASH_NOLOCK(server, hash)                                    \
  HASH_GET((server)->nss_table, struct elton_rpc_ns, _hash, nsid, hash)
// Get elton_rpc_ns by hash value of nsid.
// MUST NOT call when acquired the server->nss_table_lock.
#define GET_NS_BY_HASH(server, hash)                                           \
  ({                                                                           \
    struct elton_rpc_ns *entry;                                                \
    spin_lock(&(server)->nss_table_lock);                                      \
    entry =                                                                    \
        HASH_GET((server)->nss_table, struct elton_rpc_ns, _hash, nsid, hash); \
    spin_unlock(&(server)->nss_table_lock);                                    \
    entry;                                                                     \
  })

void elton_rpc_session_init(struct elton_rpc_session *s,
                            struct elton_rpc_server *server, u8 sid,
                            struct socket *sock);
int rpc_session_worker(void *_s);
void elton_rpc_ns_init(struct elton_rpc_ns *ns, struct elton_rpc_session *s,
                       u64 nsid, bool is_client);
u64 get_nsid(void);
u64 get_nsid_hash(struct elton_rpc_ns *ns);
u64 get_nsid_hash_by_values(u8 session_id, u64 nsid);

int rpc_sock_read_packet(struct socket *sock, u64 struct_id, void **out);
int rpc_sock_read_raw_packet(struct socket *sock, struct raw_packet **out);
int rpc_sock_write_raw_packet(struct socket *sock, struct raw_packet *raw);
