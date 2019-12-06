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

struct tree_info;

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
  struct tree_info *tree;      // FieldID=5

  // Embeds array at the tail of this struct.
  char __embeded_buffer;
  // WARNING: MUST NOT DEFINE ANY FIELD AFTER THE __embeded_buffer FIELD.
};

#define TREE_INFO_ID 8
struct tree_info {
  // Root node of directory tree.
  struct eltonfs_inode *root;
  // All inodes (using radix tree).
  //
  // Key: eltonfs_ino  (Internal inode number)
  // Value: struct eltonfs_inode *
  struct radix_tree_root *inodes;

  // Note: Original tree_info structure has two fields.  In kernel module, it is
  // difficult to built hash maps.  So we directly encode/decode to/from
  // internal representation.
};

#define GET_OBJECT_REQUEST_ID 9
struct get_object_request {
  char *id;   // FieldID=1
  u64 offset; // FieldID=2
  u64 size;   // FieldID=3
};

#define GET_OBJECT_RESPONSE_ID 10
struct get_object_response {      // StructID=10
  char *id;                       // FieldID=1
  struct elton_object_body *body; // FieldID=3

  // Embeds array at the tail of this struct.
  char __embeded_buffer;
  // WARNING: MUST NOT DEFINE ANY FIELD AFTER THE __embeded_buffer FIELD.
};

#define CREATE_OBJECT_REQUEST_ID 11
struct create_object_request {
  struct elton_object_body *body; // FieldID=1
};

#define CREATE_OBJECT_RESPONSE_ID 12
struct create_object_response {
  char *object_id; // FieldID=1

  // Embeds array at the tail of this struct.
  char __embeded_buffer;
  // WARNING: MUST NOT DEFINE ANY FIELD AFTER THE __embeded_buffer FIELD.
};

#define CREATE_COMMIT_REQUEST_ID 13
struct create_commit_request { // StructID=13
  struct commit_info info;     // FieldID=1
  char *base_commit_id;        // FieldID=2
};

#define CREATE_COMMIT_RESPONSE_ID 14
struct create_commit_response {
  char *commit_id; // FieldID=1

  // Embeds array at the tail of this struct.
  char __embeded_buffer;
  // WARNING: MUST NOT DEFINE ANY FIELD AFTER THE __embeded_buffer FIELD.
};

#define NOTIFY_LATEST_COMMIT_ID 15
struct notify_latest_commit {
  char *commit_id; // FieldID=1

  // Embeds array at the tail of this struct.
  char __embeded_buffer;
  // WARNING: MUST NOT DEFINE ANY FIELD AFTER THE __embeded_buffer FIELD.
};

#define ELTONFS_INODE_ID 18
// See struct eltonfs_inode in <elton/elton.h>

#endif // _ELTON_RPC_STRUCT_H
