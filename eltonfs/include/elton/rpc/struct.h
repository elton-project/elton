#ifndef _ELTON_RPC_STRUCT_H
#define _ELTON_RPC_STRUCT_H

#include <linux/types.h>

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
  u64 error_id;
  char *reason;

  // Embeds strings at the tail of this struct.
  char __embeded_buffer;
  // WARNING: MUST NOT DEFINE ANY FIELD AFTER THE __embeded_buffer FIELD.
};

#endif // _ELTON_RPC_STRUCT_H