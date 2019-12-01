#ifndef _ELTON_RPC_STRUCT_H
#define _ELTON_RPC_STRUCT_H

#include <elton/elton.h>
#include <linux/types.h>

struct timestamp {
  // The elapsed time from UNIX epoch.  Represents in seconds.
  u64 sec;
  // Below the decimal point of elapsed time from UNIX epoch.  Represents in
  // nanoseconds.
  u64 nsec;
};

struct tree_info {
  // Root node of directory tree.
  struct eltonfs_inode *root;
};

#define ELTON_RPC_SETUP1_ID 1
struct elton_rpc_setup1 {
  char *client_name;    // FieldID=1
  u64 version_major;    // FieldID=2
  u64 version_minor;    // FieldID=3
  u64 version_revision; // FieldID=4

  // Embeds string at the tail of this struct.
  char __embeded_buffer;
  // WARNING: MUST NOT DEFINE ANY FIELD AFTER THE __embeded_buffer FIELD.
};

#define ELTON_RPC_SETUP2_ID 2
struct elton_rpc_setup2 {
  u64 error;            // FieldID=1
  char *reason;         // FieldID=2
  char *server_name;    // FieldID=3
  u64 version_major;    // FieldID=4
  u64 version_minor;    // FieldID=5
  u64 version_revision; // FieldID=6

  // Embeds strings at the tail of this struct.
  char __embeded_buffer;
  // WARNING: MUST NOT DEFINE ANY FIELD AFTER THE __embeded_buffer FIELD.
};

#define ELTON_RPC_PING_ID 3
struct elton_rpc_ping {};

#define ELTON_RPC_ERROR_ID 4
struct elton_rpc_error {
  u64 error_id; // FieldID=1
  char *reason; // FieldID=2

  // Embeds strings at the tail of this struct.
  char __embeded_buffer;
  // WARNING: MUST NOT DEFINE ANY FIELD AFTER THE __embeded_buffer FIELD.
};

#define ELTON_OBJECT_INFO_ID 5
struct elton_object_info {
  size_t hash_length;
  u8 *hash;                    // FieldID=1
  char *hash_algorithm;        // FieldID=2
  struct timestamp created_at; // FieldID=3
  u64 size;                    // FieldID=4

  // Embeds array and strings at the tail of this struct.
  char __embeded_buffer;
  // WARNING: MUST NOT DEFINE ANY FIELD AFTER THE __embeded_buffer FIELD.
};

#define ELTON_OBJECT_BODY_ID 6
struct elton_object_body {
  size_t contents_length;
  u8 *contents; // FieldID=1
  u64 offset;   // FieldID=2

  // Embeds array at the tail of this struct.
  char __embeded_buffer;
  // WARNING: MUST NOT DEFINE ANY FIELD AFTER THE __embeded_buffer FIELD.
};

#define COMMIT_INFO_ID 7
struct commit_info {
  struct timestamp created_at; // FieldID=1
  char *left_parent_id;        // FieldID=2
  char *right_parent_id;       // FieldID=3
  struct tree_info tree;       // FieldID=5

  // Embeds array at the tail of this struct.
  char __embeded_buffer;
  // WARNING: MUST NOT DEFINE ANY FIELD AFTER THE __embeded_buffer FIELD.
};

#endif // _ELTON_RPC_STRUCT_H
