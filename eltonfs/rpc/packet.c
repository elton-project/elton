// This file implements encoders and decoders for the eltonfs-rpc.
//
// How to add new encoder and decoder.
//   1. Define a struct and constant value (macro) in elton/struct.h
//   2. Define helper functions of encoder and decoder using macros.
//
//      DECODER_DATA(eltonfs_inode) { ... };
//      IMPL_ENCODER(eltonfs_inode) {
//          int error;
//          ...
//          return 0;
//      }
//      IMPL_DECODER_PREPARE(eltonfs_inode) {
//          int error;
//          ...
//          return 0;
//      }
//      IMPL_DECODER_BODY(eltonfs_inode) {
//          int error;
//          ...
//          return 0;
//      }
//      DEFINE_ENCDEC(eltonfs_inode, ELTON_FILE_ID);
//
//   3. Register XXX_entry to look_table.
#define ELTON_LOG_PREFIX "[rpc/packet] "

#include <elton/assert.h>
#include <elton/compiler_attributes.h>
#include <elton/error.h>
#include <elton/rpc/packet.h>
#include <elton/rpc/struct.h>
#include <elton/xdr/interface.h>
#include <linux/bug.h>
#include <linux/vmalloc.h>

// Encodes packet and returns a pointer to raw_packet.
//
// Arguments:
//   struct_id:      Struct ID.
//   struct_type:    Type name of the target struct.
//   in:             Variable name of struct packet.
//   encode_process: A statement or block that encodes a struct.
//
// Encode Process:
//   The encode_process can use following variables.
//     - struct xdr_encoder enc
//     - struct xdr_struct_encoder se
//     - struct_type *s
//
// Returns: struct raw_packet *
//   The session_id and flags field ARE NOT initialized.  MUST set these fields
//   on callee.
#define ENCODE(struct_id_, struct_type, in, encode_process)                    \
  ({                                                                           \
    struct_type *s;                                                            \
    struct xdr_encoder enc;                                                    \
    struct xdr_struct_encoder se;                                              \
    struct raw_packet *raw = NULL;                                             \
    int error = 0;                                                             \
                                                                               \
    BUG_ON(in == NULL);                                                        \
    BUG_ON(in->struct_id != (struct_id_));                                     \
    BUG_ON(in->data == NULL);                                                  \
                                                                               \
    s = (struct_type *)in->data;                                               \
    GOTO_IF(error, default_encoder_init(&enc, NULL, 0));                       \
    do {                                                                       \
      size_t size;                                                             \
      /* The behavior is different between the first and second time.          \
       *                                                                       \
       * First time: calculate the required buffer size.                       \
       * Second time: encode data to buffer. */                                \
      se.enc = NULL; /* Initialize sd.enc to check the se is used or not. */   \
      encode_process;                                                          \
      if (enc.error)                                                           \
        GOTO_IF(error, enc.error);                                             \
      if (se.enc != NULL && !se.op->is_closed(&se)) {                          \
        /* Encode_process used the se.  But se is not closed. */               \
        ERR("ENCODE: 'se' is not closed.  Must check the logic in "            \
            "encode_process.");                                                \
        BUG();                                                                 \
      }                                                                        \
                                                                               \
      /* Break the loop when second time. */                                   \
      if (enc.buffer)                                                          \
        goto finish;                                                           \
                                                                               \
      /* Initialize raw_packet. */                                             \
      size = enc.pos;                                                          \
      raw = (struct raw_packet *)vmalloc(sizeof(struct raw_packet) + size);    \
      if (raw == NULL)                                                         \
        GOTO_IF(error, -ENOMEM);                                               \
      raw->size = size;                                                        \
      raw->struct_id = in->struct_id;                                          \
      raw->free = free_raw_packet;                                             \
      raw->data = &raw->__embeded_buffer;                                      \
                                                                               \
      /* Set buffer to encoder and start the second time loop. */              \
      GOTO_IF(error, default_encoder_init(&enc, raw->data, raw->size));        \
    } while (1);                                                               \
                                                                               \
  error:                                                                       \
    if (raw)                                                                   \
      vfree(raw);                                                              \
    return error;                                                              \
                                                                               \
  finish:                                                                      \
    BUG_ON(raw == NULL);                                                       \
    BUG_ON(raw->size != enc.pos);                                              \
    raw;                                                                       \
  })

// Decodes packet from raw_packet.
//
// Arguments:
//   struct_type:      Type name of the target struct.
//   in:               Variable name of struct raw_packet.
//   additional_space: The expression or statement expression to calculate
//                     additional space of struct_type.
//   decode_process:   Statements or block that decodes a struct.
//
// Decode Process:
//   The deocde_process can use following variables.
//     - struct xdr_decoder dec
//     - struct xdr_struct_decoder sd
//     - struct_type *s
//
// Returns: struct_type *
//   Returns the pointer to the specified type.  Should release it after used
//   with elton_rpc_free_decoded_data().
#define DECODE(struct_id_, struct_type, in, additional_space, decode_process)  \
  ({                                                                           \
    struct xdr_decoder dec;                                                    \
    struct xdr_struct_decoder sd;                                              \
    size_t size;                                                               \
    struct_type *s;                                                            \
    int error = 0;                                                             \
                                                                               \
    BUG_ON(in == NULL);                                                        \
    BUG_ON(in->struct_id != (struct_id_));                                     \
    BUG_ON(in->data == NULL);                                                  \
                                                                               \
    /* Calculate additional space of struct_type. */                           \
    GOTO_IF(error, default_decoder_init(&dec, in->data, in->size));            \
    size = additional_space;                                                   \
    if (dec.error)                                                             \
      GOTO_IF(error, dec.error);                                               \
                                                                               \
    /* Allocate memory of strct_type. */                                       \
    s = (struct_type *)kmalloc(sizeof(struct_type) + size, GFP_KERNEL);        \
    if (s == NULL)                                                             \
      GOTO_IF(error, -ENOMEM);                                                 \
                                                                               \
    /* Decode it. */                                                           \
    GOTO_IF(error_s, default_decoder_init(&dec, in->data, in->size));          \
    sd.dec = NULL; /* Initialize sd.dec to check the sd is used or not. */     \
    decode_process;                                                            \
    if (dec.error)                                                             \
      GOTO_IF(error_s, dec.error);                                             \
    if (sd.dec != NULL && !sd.op->is_closed(&sd)) {                            \
      /* Decode_process used the sd.  But sd is not closed. */                 \
      ERR("DECODE: 'sd' is not closed.  Must check the logic in "              \
          "decode_process.");                                                  \
      BUG();                                                                   \
    }                                                                          \
    goto finish;                                                               \
                                                                               \
  error_s:                                                                     \
    kfree(s);                                                                  \
  error:                                                                       \
    return error;                                                              \
                                                                               \
  finish:                                                                      \
    s;                                                                         \
  })

typedef int (*prepare_fn)(struct xdr_decoder *dec,
                          struct xdr_struct_decoder *sd, size_t *size,
                          void *data);
typedef int (*decode_fn)(struct xdr_decoder *dec, struct xdr_struct_decoder *sd,
                         struct raw_packet *in, void *data);
static inline int __DECODE_WITH(struct xdr_decoder *dec, void **out,
                                size_t struct_size, void *data,
                                prepare_fn calc_size, decode_fn decode) {
  int error;
  struct xdr_struct_decoder _sd;
  struct xdr_struct_decoder *sd = &_sd;
  size_t additional_size;
  void *s;
  struct xdr_decoder dec_backup;

  // Backup decoder status.
  memcpy(&dec_backup, dec, sizeof(*dec));

  RETURN_IF(calc_size(dec, sd, &additional_size, data));
  RETURN_IF(dec->error);

  // Allocate memory of strct_type.
  s = kmalloc(struct_size + additional_size, GFP_KERNEL);
  if (s == NULL)
    RETURN_IF(-ENOMEM);

  // Restore decoder status.
  memcpy(dec, &dec_backup, sizeof(*dec));

  sd->dec = NULL; // Initialize sd->dec to check the sd is used or not.
  GOTO_IF(error, decode(dec, sd, s, data));
  GOTO_IF(error, dec->error);
  if (sd->dec != NULL && !sd->op->is_closed(sd)) {
    /* Decode_process used the sd.  But sd is not closed. */
    ERR("DECODE: 'sd' is not closed.  Must check the logic in "
        "decode_process.");
    BUG();
  }
  *out = s;
  return 0;

error:
  kfree(s);
  return error;
}
static inline int __DECODE(u64 struct_id, struct raw_packet *in, void **out,
                           size_t struct_size, void *data, prepare_fn calc_size,
                           decode_fn decode) {
  int error;
  struct xdr_decoder dec;

  BUG_ON(in == NULL);
  BUG_ON(in->struct_id != struct_id);
  BUG_ON(in->data == NULL);

  RETURN_IF(default_decoder_init(&dec, in->data, in->size));
  RETURN_IF(__DECODE_WITH(&dec, out, struct_size, data, calc_size, decode));
  return 0;
}

#define __DECLARE_ENCODER_WITH(type_name)                                      \
  __unused static inline int type_name##_encode_with(                          \
      struct xdr_encoder *enc, struct xdr_struct_encoder *se,                  \
      struct type_name *s)
#define IMPL_ENCODER(type_name) __DECLARE_ENCODER_WITH(type_name)
#define __DECLARE_ENCODER(type_name)                                           \
  static int type_name##_encode(struct packet *in, struct raw_packet **out)
#define __DEFINE_ENCODER(type_name, struct_id)                                 \
  __DECLARE_ENCODER(type_name) {                                               \
    *out = ENCODE(struct_id, struct type_name, in,                             \
                  error = type_name##_encode_with(&enc, &se, s));              \
    return 0;                                                                  \
  }
#define CALL_ENCODER(type_name, enc, s)                                        \
  ({                                                                           \
    struct xdr_struct_encoder se2;                                             \
    typecheck(struct type_name *, (s));                                        \
    type_name##_encode_with((enc), &se2, (s));                                 \
  })
#define DECODER_DATA(type_name) struct __##type_name##_decoder_data
#define IMPL_DECODER_PREPARE(type_name)                                        \
  static inline int __##type_name##_decode_pre(                                \
      struct xdr_decoder *dec, struct xdr_struct_decoder *sd, size_t *size,    \
      DECODER_DATA(type_name) * data)
#define IMPL_DECODER_BODY(type_name)                                           \
  static inline int __##type_name##_decode_body(                               \
      struct xdr_decoder *dec, struct xdr_struct_decoder *sd,                  \
      struct type_name *s, DECODER_DATA(type_name) * data)
#define __DECLARE_DECODER(type_name)                                           \
  static int type_name##_decode(struct raw_packet *in, void **out)
#define __DEFINE_DECODER(type_name, struct_id)                                 \
  __DECLARE_DECODER(type_name) {                                               \
    int error;                                                                 \
    DECODER_DATA(type_name) data = {};                                         \
    RETURN_IF(__DECODE(struct_id, in, out, sizeof(struct type_name), &data,    \
                       (void *)__##type_name##_decode_pre,                     \
                       (void *)__##type_name##_decode_body));                  \
    return 0;                                                                  \
  }
#define __DECLARE_DECODER_WITH(type_name)                                      \
  __unused static int type_name##_decode_with(struct xdr_decoder *dec,         \
                                              void **out)
#define __DEFINE_DECODER_WITH(type_name, struct_id)                            \
  __DECLARE_DECODER_WITH(type_name) {                                          \
    int error;                                                                 \
    DECODER_DATA(type_name) data = {};                                         \
    RETURN_IF(__DECODE_WITH(dec, out, sizeof(struct type_name), &data,         \
                            (void *)__##type_name##_decode_pre,                \
                            (void *)__##type_name##_decode_body));             \
    return 0;                                                                  \
  }
#define DECLARE_ENCDEC(type_name)                                              \
  __DECLARE_ENCODER(type_name);                                                \
  __DECLARE_ENCODER_WITH(type_name);                                           \
  __DECLARE_DECODER(type_name);                                                \
  __DECLARE_DECODER_WITH(type_name);
#define DEFINE_ENCDEC(type_name, struct_id)                                    \
  __DEFINE_ENCODER(type_name, struct_id);                                      \
  __DEFINE_DECODER(type_name, struct_id);                                      \
  __DEFINE_DECODER_WITH(type_name, struct_id);                                 \
  const static struct entry type_name##_entry = {                              \
      .encode = type_name##_encode,                                            \
      .decode = type_name##_decode,                                            \
  }
#define DEFINE_ENC_ONLY(type_name, struct_id)                                  \
  __DEFINE_ENCODER(type_name, struct_id);                                      \
  const static struct entry type_name##_entry = {                              \
      .encode = type_name##_encode,                                            \
      .decode = not_implemented_decode,                                        \
  }
#define DEFINE_DEC_ONLY(type_name, struct_id)                                  \
  __DEFINE_DECODER(type_name, struct_id);                                      \
  __DEFINE_DECODER_WITH(type_name, struct_id);                                 \
  const static struct entry type_name##_entry = {                              \
      .encode = not_implemented_encode,                                        \
      .decode = type_name##_decode,                                            \
  }
#define CALL_DECODER(type_name, dec, out)                                      \
  ({                                                                           \
    typecheck(struct type_name **, out);                                       \
    type_name##_decode_with((dec), (void **)(out));                            \
  })

void free_raw_packet(struct raw_packet *packet) { vfree(packet); }

struct entry {
  int (*encode)(struct packet *in, struct raw_packet **out);
  int (*decode)(struct raw_packet *in, void **out);
};

static int not_implemented_encode(struct packet *in, struct raw_packet **out) {
  ERR("not implemented encoder: struct_id=%d", in->struct_id);
  BUG();
  // Unreachable.
  return 0;
}
static int not_implemented_decode(struct raw_packet *in, void **out) {
  ERR("not implemented decoder: struct_id=%llu", in->struct_id);
  BUG();
  // Unreachable.
  return 0;
}

DECLARE_ENCDEC(eltonfs_inode_xdr);

static int setup1_decode(struct raw_packet *in, void **out) {
  size_t str_size = 0;
  *out = DECODE(
      ELTON_RPC_SETUP1_ID, struct elton_rpc_setup1, in, ({
        do {
          BREAK_IF(dec.dec_op->struct_(&dec, &sd));
          BREAK_IF(sd.op->bytes(&sd, 1, NULL, &str_size)); // Field 1
        } while (0);
        str_size + 1;
      }),
      ({
        do {
          // Initialize setup1.
          s->client_name = &s->__embeded_buffer;
          // Decodes.
          BREAK_IF(dec.dec_op->struct_(&dec, &sd));
          BREAK_IF(sd.op->bytes(&sd, 1, s->client_name, &str_size)); // Field 1
          s->client_name[str_size] = '\0';
          BREAK_IF(sd.op->u64(&sd, 2, &s->version_major));    // Field 2
          BREAK_IF(sd.op->u64(&sd, 3, &s->version_minor));    // Field 3
          BREAK_IF(sd.op->u64(&sd, 4, &s->version_revision)); // Field 4
          BREAK_IF(sd.op->close(&sd));
        } while (0);
      }));
  return 0;
}
const static struct entry setup1_entry = {
    .encode = not_implemented_encode,
    .decode = setup1_decode,
};

static int setup2_encode(struct packet *in, struct raw_packet **out) {
  *out =
      ENCODE(ELTON_RPC_SETUP2_ID, struct elton_rpc_setup2, in, ({
               do {
                 BREAK_IF(enc.enc_op->struct_(&enc, &se, 6));
                 BREAK_IF(se.op->u64(&se, 1, s->error)); // Field 1
                 BREAK_IF(se.op->bytes(&se, 2, s->reason,
                                       strlen(s->reason))); // Field 2
                 BREAK_IF(se.op->bytes(&se, 3, s->server_name,
                                       strlen(s->server_name)));    // Field 3
                 BREAK_IF(se.op->u64(&se, 4, s->version_major));    // Field 4
                 BREAK_IF(se.op->u64(&se, 5, s->version_minor));    // Field 5
                 BREAK_IF(se.op->u64(&se, 6, s->version_revision)); // Field 6
                 BREAK_IF(se.op->close(&se));
               } while (0);
             }));
  return 0;
}
const static struct entry setup2_entry = {
    .encode = setup2_encode,
    .decode = not_implemented_decode,
};

static int ping_encode(struct packet *in, struct raw_packet **out) {
  *out = ENCODE(ELTON_RPC_PING_ID, struct elton_rpc_ping, in, ({
                  do {
                    BREAK_IF(enc.enc_op->struct_(&enc, &se, 0));
                    BREAK_IF(se.op->close(&se));
                  } while (0);
                }));
  return 0;
}
static int ping_decode(struct raw_packet *in, void **out) {
  *out = DECODE(ELTON_RPC_PING_ID, struct elton_rpc_ping, in, 0, ({
                  do {
                    BREAK_IF(dec.dec_op->struct_(&dec, &sd));
                    BREAK_IF(sd.op->close(&sd));
                  } while (0);
                }));
  return 0;
}
const static struct entry ping_entry = {
    .encode = ping_encode,
    .decode = ping_decode,
};

static int error_encode(struct packet *in, struct raw_packet **out) {
  *out = ENCODE(ELTON_RPC_ERROR_ID, struct elton_rpc_error, in, ({
                  do {
                    BREAK_IF(enc.enc_op->struct_(&enc, &se, 2));
                    BREAK_IF(se.op->u64(&se, 1, s->error_id)); // Field 1
                    BREAK_IF(se.op->bytes(&se, 2, s->reason,
                                          strlen(s->reason))); // Field 2
                    BREAK_IF(se.op->close(&se));
                  } while (0);
                }));
  return 0;
}
static int error_decode(struct raw_packet *in, void **out) {
  size_t reason_size = 0;
  *out = DECODE(
      ELTON_RPC_ERROR_ID, struct elton_rpc_error, in, ({
        do {
          u64 dummy;
          BREAK_IF(dec.dec_op->struct_(&dec, &sd));
          BREAK_IF(sd.op->u64(&sd, 1, &dummy));
          BREAK_IF(sd.op->bytes(&sd, 2, NULL, &reason_size));
        } while (0);
        reason_size + 1;
      }),
      ({
        do {
          // Initialize error.
          s->reason = &s->__embeded_buffer;

          // Decode
          BREAK_IF(dec.dec_op->struct_(&dec, &sd));
          BREAK_IF(sd.op->u64(&sd, 1, &s->error_id));              // Field 1
          BREAK_IF(sd.op->bytes(&sd, 2, s->reason, &reason_size)); // Field 2
          s->reason[reason_size] = '\0';
          BREAK_IF(sd.op->close(&sd));
        } while (0);
      }));
  return 0;
}
const static struct entry error_entry = {
    .encode = error_encode,
    .decode = error_decode,
};

static int elton_object_info_encode(struct packet *in,
                                    struct raw_packet **out) {
  *out = ENCODE(ELTON_OBJECT_INFO_ID, struct elton_object_info, in, ({
                  do {
                    BREAK_IF(enc.enc_op->struct_(&enc, &se, 4));
                    BREAK_IF(se.op->bytes(&se, 1, s->hash, s->hash_length));
                    BREAK_IF(se.op->bytes(&se, 2, s->hash_algorithm,
                                          strlen(s->hash_algorithm)));
                    BREAK_IF(se.op->timestamp(&se, 3, s->created_at));
                    BREAK_IF(se.op->u64(&se, 4, s->size));
                    BREAK_IF(se.op->close(&se));
                  } while (0);
                }));
  return 0;
}
static int elton_object_info_decode(struct raw_packet *in, void **out) {
  size_t hash_length = 0;
  size_t algo_length = 0;
  *out = DECODE(ELTON_OBJECT_INFO_ID, struct elton_object_info, in, ({
                  do {
                    BREAK_IF(dec.dec_op->struct_(&dec, &sd));
                    BREAK_IF(sd.op->bytes(&sd, 1, NULL, &hash_length));
                    BREAK_IF(sd.op->bytes(&sd, 2, NULL, &algo_length));
                  } while (0);
                  // Increase buffer size to allocate NULL element at the end
                  // of array.
                  hash_length + algo_length + 1;
                }),
                ({
                  do {
                    // Initialize arrays.
                    s->hash = &s->__embeded_buffer;
                    s->hash_algorithm = &s->__embeded_buffer + hash_length;

                    // Decode
                    BREAK_IF(dec.dec_op->struct_(&dec, &sd));
                    s->hash_length = hash_length;
                    BREAK_IF(sd.op->bytes(&sd, 1, s->hash, &s->hash_length));
                    BREAK_IF(
                        sd.op->bytes(&sd, 2, s->hash_algorithm, &algo_length));
                    s->hash_algorithm[algo_length] = '\0';
                    BREAK_IF(sd.op->timestamp(&sd, 3, &s->created_at));
                    BREAK_IF(sd.op->u64(&sd, 4, &s->size));
                    BREAK_IF(sd.op->close(&sd));
                  } while (0);
                }));
  return 0;
}
const static struct entry elton_object_info_entry = {
    .encode = elton_object_info_encode,
    .decode = elton_object_info_decode,
};

DECODER_DATA(elton_object_body) { size_t contents_length; };
IMPL_ENCODER(elton_object_body) {
  int error;
  RETURN_IF(enc->enc_op->struct_(enc, se, 2));
  RETURN_IF(se->op->bytes(se, 1, s->contents, s->contents_length));
  RETURN_IF(se->op->u64(se, 2, s->offset));
  RETURN_IF(se->op->close(se));
  return 0;
}
IMPL_DECODER_PREPARE(elton_object_body) {
  int error;
  RETURN_IF(dec->dec_op->struct_(dec, sd));
  RETURN_IF(sd->op->bytes(sd, 1, NULL, &data->contents_length));
  *size = data->contents_length + 1;
  return 0;
}
IMPL_DECODER_BODY(elton_object_body) {
  int error;
  // Initialize error.
  s->contents = &s->__embeded_buffer;

  // Decode
  RETURN_IF(dec->dec_op->struct_(dec, sd));
  s->contents_length = data->contents_length;
  RETURN_IF(sd->op->bytes(sd, 1, s->contents, &s->contents_length));
  s->contents[s->contents_length] = '\0';
  RETURN_IF(sd->op->u64(sd, 2, &s->offset));
  RETURN_IF(sd->op->close(sd));
  return 0;
}
DEFINE_ENCDEC(elton_object_body, ELTON_OBJECT_BODY_ID);

DECLARE_ENCDEC(tree_info);
DECODER_DATA(commit_info) {
  size_t left_length;
  size_t right_length;
};
IMPL_ENCODER(commit_info) {
  int error;
  RETURN_IF(enc->enc_op->struct_(enc, se, 4));
  RETURN_IF(se->op->timestamp(se, 1, s->created_at));
  RETURN_IF(se->op->bytes(se, 2, s->left_parent_id, strlen(s->left_parent_id)));
  RETURN_IF(
      se->op->bytes(se, 3, s->right_parent_id, strlen(s->right_parent_id)));
  RETURN_IF(se->op->external_encoder(se, 5));
  RETURN_IF(CALL_ENCODER(tree_info, enc, s->tree));
  RETURN_IF(se->op->close(se));
  return 0;
}
IMPL_DECODER_PREPARE(commit_info) {
  int error;
  RETURN_IF(dec->dec_op->struct_(dec, sd));
  RETURN_IF(sd->op->timestamp(sd, 1, NULL));
  RETURN_IF(sd->op->bytes(sd, 2, NULL, &data->left_length));
  RETURN_IF(sd->op->bytes(sd, 3, NULL, &data->right_length));
  *size = data->left_length + data->right_length + 2;
  return 0;
}
IMPL_DECODER_BODY(commit_info) {
  int error;
  // Initialize error.
  s->left_parent_id = &s->__embeded_buffer;
  s->right_parent_id = &s->left_parent_id[data->left_length + 1];

  // Decode
  RETURN_IF(dec->dec_op->struct_(dec, sd));
  RETURN_IF(sd->op->timestamp(sd, 1, &s->created_at));
  RETURN_IF(sd->op->bytes(sd, 2, s->left_parent_id, &data->left_length));
  s->left_parent_id[data->left_length] = '\0';
  RETURN_IF(sd->op->bytes(sd, 3, s->right_parent_id, &data->right_length));
  s->right_parent_id[data->right_length] = '\0';
  RETURN_IF(sd->op->external_decoder(sd, 5));
  RETURN_IF(CALL_DECODER(tree_info, dec, &s->tree));
  RETURN_IF(sd->op->close(sd));
  return 0;
}
DEFINE_ENCDEC(commit_info, COMMIT_INFO_ID);

DECODER_DATA(tree_info){};
IMPL_ENCODER(tree_info) {
  int error;
  struct xdr_map_encoder _me;
  struct xdr_map_encoder *me = &_me;
  void **slot;
  struct radix_tree_iter iter;
  u64 inode_count = 0;

  radix_tree_for_each_slot(slot, s->inodes, &iter, 0) inode_count++;

  RETURN_IF(enc->enc_op->struct_(enc, se, 2));
  RETURN_IF(se->op->u64(se, 3, s->root->eltonfs_ino));
  RETURN_IF(se->op->map(se, 4, me, inode_count));
  radix_tree_for_each_slot(slot, s->inodes, &iter, 0) {
    struct eltonfs_inode_xdr *inode = (struct eltonfs_inode_xdr *)*slot;

    RETURN_IF(enc->enc_op->u64(enc, inode->eltonfs_ino));
    RETURN_IF(CALL_ENCODER(eltonfs_inode_xdr, enc, inode));
    RETURN_IF(me->op->encoded_kv(me));
  }
  RETURN_IF(me->op->close(me));
  RETURN_IF(se->op->close(se));
  return 0;
}
IMPL_DECODER_PREPARE(tree_info) {
  *size = 0;
  return 0;
}
IMPL_DECODER_BODY(tree_info) {
  int error;
  u64 root_ino;
  struct xdr_map_decoder _mdec;
  struct xdr_map_decoder *mdec = &_mdec;

  s->inodes = (struct radix_tree_root *)kmalloc(sizeof(*s->inodes), GFP_KERNEL);
  if (s->inodes == NULL)
    // TODO: remove radix tree when an error occured.
    RETURN_IF(-ENOMEM);
  INIT_RADIX_TREE(s->inodes, GFP_KERNEL);

  RETURN_IF(dec->dec_op->struct_(dec, sd));
  RETURN_IF(sd->op->u64(sd, 3, &root_ino));
  RETURN_IF(sd->op->map(sd, 4, mdec));
  while (mdec->op->has_next_kv(mdec)) {
    u64 ino;
    struct eltonfs_inode_xdr *inode;

    RETURN_IF(dec->dec_op->u64(dec, &ino));
    RETURN_IF(CALL_DECODER(eltonfs_inode_xdr, dec, &inode));
    RETURN_IF(mdec->op->decoded_kv(mdec));
    inode->eltonfs_ino = ino;

    RETURN_IF(radix_tree_insert(s->inodes, ino, inode));
  }
  RETURN_IF(mdec->op->close(mdec));
  RETURN_IF(sd->op->close(sd));
  s->root = radix_tree_lookup(s->inodes, root_ino);
  WARN_ONCE(!s->root, "not found root inode");
  return 0;
}
DEFINE_ENCDEC(tree_info, TREE_INFO_ID);

IMPL_ENCODER(get_object_request) {
  int error;
  RETURN_IF(enc->enc_op->struct_(enc, se, 3));
  RETURN_IF(se->op->bytes(se, 1, s->id, strlen(s->id)));
  RETURN_IF(se->op->u64(se, 2, s->offset));
  RETURN_IF(se->op->u64(se, 3, s->size));
  RETURN_IF(se->op->close(se));
  return 0;
}
DEFINE_ENC_ONLY(get_object_request, GET_OBJECT_REQUEST_ID);

DECODER_DATA(get_object_response) { size_t id_length; };
IMPL_DECODER_PREPARE(get_object_response) {
  int error;
  RETURN_IF(dec->dec_op->struct_(dec, sd));
  RETURN_IF(sd->op->bytes(sd, 1, NULL, &data->id_length));
  *size = data->id_length + 1;
  return 0;
}
IMPL_DECODER_BODY(get_object_response) {
  int error;
  s->id = &s->__embeded_buffer;

  RETURN_IF(dec->dec_op->struct_(dec, sd));
  RETURN_IF(sd->op->bytes(sd, 1, s->id, &data->id_length));
  s->id[data->id_length] = '\0';
  RETURN_IF(sd->op->external_decoder(sd, 3));
  RETURN_IF(CALL_DECODER(elton_object_body, dec, &s->body));
  RETURN_IF(sd->op->close(sd));
  return 0;
}
DEFINE_DEC_ONLY(get_object_response, GET_OBJECT_RESPONSE_ID);

IMPL_ENCODER(create_object_request) {
  int error;
  RETURN_IF(enc->enc_op->struct_(enc, se, 1));
  RETURN_IF(se->op->external_encoder(se, 1));
  RETURN_IF(CALL_ENCODER(elton_object_body, enc, s->body));
  RETURN_IF(se->op->close(se));
  return 0;
}
DEFINE_ENC_ONLY(create_object_request, CREATE_OBJECT_REQUEST_ID);

DECODER_DATA(create_object_response) { size_t id_length; };
IMPL_DECODER_PREPARE(create_object_response) {
  int error;
  RETURN_IF(dec->dec_op->struct_(dec, sd));
  RETURN_IF(sd->op->bytes(sd, 1, NULL, &data->id_length));
  *size = data->id_length + 1;
  return 0;
}
IMPL_DECODER_BODY(create_object_response) {
  int error;
  s->object_id = &s->__embeded_buffer;

  RETURN_IF(dec->dec_op->struct_(dec, sd));
  RETURN_IF(sd->op->bytes(sd, 1, s->object_id, &data->id_length));
  s->object_id[data->id_length] = '\0';
  RETURN_IF(sd->op->close(sd));
  return 0;
}
DEFINE_DEC_ONLY(create_object_response, CREATE_OBJECT_RESPONSE_ID);

IMPL_ENCODER(create_commit_request) {
  int error;
  RETURN_IF(enc->enc_op->struct_(enc, se, 1));
  RETURN_IF(se->op->external_encoder(se, 1));
  RETURN_IF(CALL_ENCODER(commit_info, enc, &s->info));
  RETURN_IF(se->op->close(se));
  return 0;
}
DEFINE_ENC_ONLY(create_commit_request, CREATE_COMMIT_REQUEST_ID);

DECODER_DATA(create_commit_response) { size_t id_length; };
IMPL_DECODER_PREPARE(create_commit_response) {
  int error;
  RETURN_IF(dec->dec_op->struct_(dec, sd));
  RETURN_IF(sd->op->bytes(sd, 1, NULL, &data->id_length));
  *size = data->id_length + 1;
  return 0;
}
IMPL_DECODER_BODY(create_commit_response) {
  int error;
  s->commit_id = &s->__embeded_buffer;

  RETURN_IF(dec->dec_op->struct_(dec, sd));
  RETURN_IF(sd->op->bytes(sd, 1, s->commit_id, &data->id_length));
  s->commit_id[data->id_length] = '\0';
  RETURN_IF(sd->op->close(sd));
  return 0;
}
DEFINE_DEC_ONLY(create_commit_response, CREATE_COMMIT_RESPONSE_ID);

DECODER_DATA(notify_latest_commit) { size_t id_length; };
IMPL_DECODER_PREPARE(notify_latest_commit) {
  int error;
  RETURN_IF(dec->dec_op->struct_(dec, sd));
  RETURN_IF(sd->op->bytes(sd, 1, NULL, &data->id_length));
  *size = data->id_length + 1;
  return 0;
}
IMPL_DECODER_BODY(notify_latest_commit) {
  int error;
  s->commit_id = &s->__embeded_buffer;

  RETURN_IF(dec->dec_op->struct_(dec, sd));
  RETURN_IF(sd->op->bytes(sd, 1, s->commit_id, &data->id_length));
  s->commit_id[data->id_length] = '\0';
  RETURN_IF(sd->op->close(sd));
  return 0;
}
DEFINE_DEC_ONLY(notify_latest_commit, NOTIFY_LATEST_COMMIT_ID);

IMPL_ENCODER(get_commit_info_request) {
  int error;
  RETURN_IF(enc->enc_op->struct_(enc, se, 1));
  RETURN_IF(se->op->bytes(se, 1, s->commit_id, strlen(s->commit_id)));
  RETURN_IF(se->op->close(se));
  return 0;
}
DEFINE_ENC_ONLY(get_commit_info_request, GET_COMMIT_INFO_REQUEST_ID);

DECODER_DATA(get_commit_info_response) { size_t id_length; };
IMPL_DECODER_PREPARE(get_commit_info_response) {
  int error;
  RETURN_IF(dec->dec_op->struct_(dec, sd));
  RETURN_IF(sd->op->bytes(sd, 1, NULL, &data->id_length));
  *size = data->id_length + 1;
  return 0;
}
IMPL_DECODER_BODY(get_commit_info_response) {
  int error;
  s->commit_id = &s->__embeded_buffer;

  RETURN_IF(dec->dec_op->struct_(dec, sd));
  RETURN_IF(sd->op->bytes(sd, 1, s->commit_id, &data->id_length));
  s->commit_id[data->id_length] = '\0';
  RETURN_IF(sd->op->external_decoder(sd, 2));
  RETURN_IF(CALL_DECODER(commit_info, dec, &s->info));
  RETURN_IF(sd->op->close(sd));
  return 0;
}
DEFINE_DEC_ONLY(get_commit_info_response, GET_COMMIT_INFO_RESPONSE_ID);

DECODER_DATA(eltonfs_inode_xdr) { size_t id_length; };
IMPL_ENCODER(eltonfs_inode_xdr) {
  int error;
  struct xdr_map_encoder _map;
  struct xdr_map_encoder *map = &_map;
  struct eltonfs_dir_entry *entry;
  RETURN_IF(enc->enc_op->struct_(enc, se, 10));

  RETURN_IF(se->op->bytes(se, 1, s->object_id, strlen(s->object_id)));
  RETURN_IF(se->op->u64(se, 3, s->mode));
  RETURN_IF(se->op->u64(se, 4, s->owner));
  RETURN_IF(se->op->u64(se, 5, s->group));
  RETURN_IF(se->op->timestamp(se, 6, s->atime));
  RETURN_IF(se->op->timestamp(se, 7, s->mtime));
  RETURN_IF(se->op->timestamp(se, 8, s->ctime));
  RETURN_IF(se->op->u64(se, 9, s->major));
  RETURN_IF(se->op->u64(se, 10, s->minor));

  // 11: entries
  RETURN_IF(se->op->map(se, 11, map, s->dir_entries_len));
  list_for_each_entry(entry, &s->dir_entries._list_head, _list_head) {
    enc->enc_op->bytes(enc, entry->name, entry->name_len);
    enc->enc_op->u64(enc, entry->ino);
    map->op->encoded_kv(map);
  }
  RETURN_IF(map->op->close(map));

  RETURN_IF(se->op->close(se));
  return 0;
}
IMPL_DECODER_PREPARE(eltonfs_inode_xdr) {
  int error;
  RETURN_IF(dec->dec_op->struct_(dec, sd));
  RETURN_IF(sd->op->bytes(sd, 1, NULL, &data->id_length));
  *size = 0;
  return 0;
}
IMPL_DECODER_BODY(eltonfs_inode_xdr) {
  int error;
  struct xdr_map_decoder _mdec;
  struct xdr_map_decoder *mdec = &_mdec;
  char *obj_id;

  obj_id = kmalloc(data->id_length + 1, GFP_KERNEL);
  if (obj_id == NULL) {
    RETURN_IF(-ENOMEM);
  }

  // Decode
  RETURN_IF(dec->dec_op->struct_(dec, sd));
  RETURN_IF(sd->op->bytes(sd, 1, obj_id, &data->id_length));
  obj_id[data->id_length] = '\0';
  s->object_id = obj_id;
  RETURN_IF(sd->op->u64(sd, 3, &s->mode));
  RETURN_IF(sd->op->u64(sd, 4, &s->owner));
  RETURN_IF(sd->op->u64(sd, 5, &s->group));
  RETURN_IF(sd->op->timestamp(sd, 6, &s->atime));
  RETURN_IF(sd->op->timestamp(sd, 7, &s->mtime));
  RETURN_IF(sd->op->timestamp(sd, 8, &s->ctime));
  RETURN_IF(sd->op->u64(sd, 9, &s->major));
  RETURN_IF(sd->op->u64(sd, 10, &s->minor));

  INIT_LIST_HEAD(&s->dir_entries._list_head);
  RETURN_IF(sd->op->map(sd, 11, mdec));
  s->dir_entries_len = 0;
  while (mdec->op->has_next_kv(mdec)) {
    size_t len;
    struct eltonfs_dir_entry *eino;
    eino = (struct eltonfs_dir_entry *)kmalloc(sizeof(*eino), GFP_KERNEL);
    if (eino == NULL) {
      RETURN_IF(-ENOMEM);
    }

    len = ELTONFS_NAME_LEN;
    RETURN_IF(dec->dec_op->bytes(dec, eino->name, &len));
    eino->name_len = len;
    eino->name[len] = '\0';
    RETURN_IF(dec->dec_op->u64(dec, &eino->ino));
    RETURN_IF(mdec->op->decoded_kv(mdec));

    list_add_tail(&eino->_list_head, &s->dir_entries._list_head);
    s->dir_entries_len++;
  }
  RETURN_IF(mdec->op->close(mdec));
  RETURN_IF(sd->op->close(sd));
  return 0;
}
DEFINE_ENCDEC(eltonfs_inode_xdr, ELTONFS_INODE_ID);

IMPL_ENCODER(notify_latest_commit_request) {
  int error;
  RETURN_IF(enc->enc_op->struct_(enc, se, 1));
  RETURN_IF(se->op->bytes(se, 1, s->volume_id, strlen(s->volume_id)));
  RETURN_IF(se->op->close(se));
  return 0;
}
DEFINE_ENC_ONLY(notify_latest_commit_request, NOTIFY_LATEST_COMMIT_REQUEST_ID);

IMPL_ENCODER(get_volume_id_request) {
  int error;
  RETURN_IF(enc->enc_op->struct_(enc, se, 1));
  RETURN_IF(se->op->bytes(se, 1, s->volume_name, strlen(s->volume_name)));
  RETURN_IF(se->op->close(se));
  return 0;
}
DEFINE_ENC_ONLY(get_volume_id_request, GET_VOLUME_ID_REQUEST_ID);

DECODER_DATA(get_volume_id_response) { size_t id_len; };
IMPL_DECODER_PREPARE(get_volume_id_response) {
  int error;
  RETURN_IF(dec->dec_op->struct_(dec, sd));
  RETURN_IF(sd->op->bytes(sd, 1, NULL, &data->id_len));
  *size = data->id_len + 1;
  return 0;
}
IMPL_DECODER_BODY(get_volume_id_response) {
  int error;
  char *buffer = &s->__embeded_buffer;
  RETURN_IF(dec->dec_op->struct_(dec, sd));
  RETURN_IF(sd->op->bytes(sd, 1, buffer, &data->id_len));
  buffer[data->id_len] = '\0';
  RETURN_IF(sd->op->close(sd));
  s->volume_id = buffer;
  return 0;
}
DEFINE_DEC_ONLY(get_volume_id_response, GET_VOLUME_ID_RESPONSE_ID);

// Lookup table from struct_id to encoder/decoder function.
const static struct entry *look_table[] = {
    // StructID 0: invalid
    NULL,
    // StructID 1: setup1
    &setup1_entry,
    // StructID 2: setup2
    &setup2_entry,
    // StructID 3: ping
    &ping_entry,
    // StructID 4: error
    &error_entry,
    // StructID 5: elton_object_info
    &elton_object_info_entry,
    // StructID 6: elton_object_body
    &elton_object_body_entry,
    // StructID 7: commit_info
    &commit_info_entry,
    // StructID 8: tree_info
    &tree_info_entry,
    // StructID 9: get_object_request
    &get_object_request_entry,
    // StructID 10: get_object_response
    &get_object_response_entry,
    // StructID 11: create_object_request
    &create_object_request_entry,
    // StructID 12: create_object_response
    &create_object_response_entry,
    // StructID 13: create_commit_request
    &create_commit_request_entry,
    // StructID 14: create_commit_response
    &create_commit_response_entry,
    // StructID 15: notify_latest_commit
    &notify_latest_commit_entry,
    // StructID 16: get_commit_info_request
    &get_commit_info_request_entry,
    // StructID 17: get_commit_info_response
    &get_commit_info_response_entry,
    // StructID 18: eltonfs_inode
    &eltonfs_inode_xdr_entry,
    // StructID 19: notify_latest_commit_request
    &notify_latest_commit_request_entry,
    // StructID 20: get_volume_id_request
    &get_volume_id_request_entry,
    // StructID 21: get_volume_id_response
    &get_volume_id_response_entry,
};
#define ELTON_MAX_STRUCT_ID 21

static int lookup(u64 struct_id, const struct entry **entry) {
  BUILD_ASSERT_EQUAL_ARRAY_SIZE(ELTON_MAX_STRUCT_ID + 1, look_table);
  BUG_ON(struct_id == 0);
  BUG_ON(struct_id > ELTON_MAX_STRUCT_ID);

  *entry = look_table[struct_id];
  return 0;
}

int elton_rpc_encode_packet(struct packet *in, struct raw_packet **out,
                            u64 session_id, u8 flags) {
  int error;
  const struct entry *entry;

  RETURN_IF(lookup(in->struct_id, &entry));
  BUG_ON(entry == NULL);
  BUG_ON(entry->encode == NULL);
  RETURN_IF(entry->encode(in, out));

  BUG_ON(*out == NULL);
  (*out)->session_id = session_id;
  (*out)->flags = flags;
  return 0;
}

int elton_rpc_decode_packet(struct raw_packet *in, void **out) {
  int error;
  const struct entry *entry;

  RETURN_IF(lookup(in->struct_id, &entry));
  BUG_ON(entry == NULL);
  BUG_ON(entry->decode == NULL);
  RETURN_IF(entry->decode(in, out));
  return 0;
}

void elton_rpc_free_decoded_data(void *data) { kfree(data); }

int elton_rpc_get_raw_packet_size(char *buff, size_t len, size_t *packet_size) {
  int error = 0;
  u64 n;
  struct xdr_decoder dec;
  RETURN_IF(default_decoder_init(&dec, buff, len));
  RETURN_IF(dec.dec_op->u64(&dec, &n));
  *packet_size = n;
  return 0;
}

int elton_rpc_build_raw_packet(struct raw_packet **out, char *buff,
                               size_t len) {
  int error = 0;
  struct xdr_decoder dec;
  struct raw_packet *raw = NULL;
  u64 data_size;

  RETURN_IF(default_decoder_init(&dec, buff, len));
  RETURN_IF(dec.dec_op->u64(&dec, &data_size));

  raw = (struct raw_packet *)vmalloc(sizeof(struct raw_packet) + data_size);
  if (raw == NULL) {
    GOTO_IF(error, -ENOMEM);
  }
  raw->size = data_size;
  GOTO_IF(error, dec.dec_op->u64(&dec, &raw->session_id));
  GOTO_IF(error, dec.dec_op->u8(&dec, &raw->flags));
  GOTO_IF(error, dec.dec_op->u64(&dec, &raw->struct_id));

  if (len < dec.pos + data_size) {
    GOTO_IF(error, -ELTON_XDR_NEED_MORE_MEM);
  }
  raw->data = &raw->__embeded_buffer;
  memcpy(raw->data, buff + dec.pos, data_size);
  raw->free = free_raw_packet;

  *out = raw;
  return 0;

error:
  if (raw)
    free_raw_packet(raw);
  return error;
}
