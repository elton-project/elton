#include <elton/assert.h>
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
//   encode_process: Statements or block that encodes a struct.
//
// Returns: struct raw_packet *
//   The session_id and flags field ARE NOT initialized.  MUST set these fields
//   on callee.
#define ENCODE(struct_id_, struct_type, in, encode_process)                    \
  ({                                                                           \
    struct_type *s;                                                            \
    struct xdr_encoder enc;                                                    \
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
      encode_process;                                                          \
      if (enc.error)                                                           \
        GOTO_IF(error, enc.error);                                             \
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

// struct_type:      Type name of the target struct.
// in:               Variable name of struct raw_packet.
// additional_space: The expression or statement expression to calculate
//                   additional space of struct_type.
// decode_process:   Statements or block that decodes a struct.
#define DECODE(struct_id_, struct_type, in, additional_space, decode_process)  \
  ({                                                                           \
    struct xdr_decoder dec;                                                    \
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
    decode_process;                                                            \
    if (dec.error)                                                             \
      GOTO_IF(error_s, dec.error);                                             \
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

void free_raw_packet(struct raw_packet *packet) { vfree(packet); }

struct entry {
  int (*encode)(struct packet *in, struct raw_packet **out);
  int (*decode)(struct raw_packet *in, void **out);
};

static int setup1_encode(struct packet *in, struct raw_packet **out) {
  ERR("setup1_encode is not implemented");
  BUG();
  return 0;
}
static int setup1_decode(struct raw_packet *in, void **out) {
  size_t str_size = 0;
  struct xdr_struct_decoder sd;
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
          sd.op->u64(&sd, 2, &s->version_major);    // Field 2
          sd.op->u64(&sd, 3, &s->version_minor);    // Field 3
          sd.op->u64(&sd, 4, &s->version_revision); // Field 4
        } while (0);
      }));
  return 0;
}
const static struct entry setup1_entry = {
    .encode = setup1_encode,
    .decode = setup1_decode,
};

static int setup2_encode(struct packet *in, struct raw_packet **out) {
  *out = ENCODE(ELTON_RPC_SETUP2_ID, struct elton_rpc_setup2, in, {
    enc.enc_op->u64(&enc, s->error);
    enc.enc_op->bytes(&enc, s->reason, strlen(s->reason));
    enc.enc_op->bytes(&enc, s->server_name, strlen(s->server_name));
    enc.enc_op->u64(&enc, s->version_major);
    enc.enc_op->u64(&enc, s->version_minor);
    enc.enc_op->u64(&enc, s->version_revision);
  });
  return 0;
}
static int setup2_decode(struct raw_packet *in, void **out) {
  size_t reason_size, name_size;
  *out = DECODE(ELTON_RPC_SETUP2_ID, struct elton_rpc_setup2, in, ({
                  u64 dummy;
                  dec.dec_op->u64(&dec, &dummy); // Skip FieldID1.
                  dec.dec_op->bytes(&dec, NULL, &reason_size);
                  dec.dec_op->bytes(&dec, NULL, &name_size);
                  reason_size + 1 + name_size + 1;
                }),
                {
                  // initialize setup2.
                  s->reason = &s->__embeded_buffer;
                  s->server_name = s->reason + reason_size + 1;
                  // Decodes.
                  dec.dec_op->u64(&dec, &s->error);
                  dec.dec_op->bytes(&dec, s->reason, &reason_size);
                  s->reason[reason_size] = '\0';
                  dec.dec_op->bytes(&dec, s->server_name, &name_size);
                  s->server_name[name_size] = '\0';
                  dec.dec_op->u64(&dec, &s->version_major);
                  dec.dec_op->u64(&dec, &s->version_minor);
                  dec.dec_op->u64(&dec, &s->version_revision);
                });
  return 0;
}
const static struct entry setup2_entry = {
    .encode = setup2_encode,
    .decode = setup2_decode,
};

static int ping_encode(struct packet *in, struct raw_packet **out) {
  *out =
      ENCODE(ELTON_RPC_PING_ID, struct elton_rpc_ping, in, {/* Do nothing. */});
  return 0;
}
static int ping_decode(struct raw_packet *in, void **out) {
  *out = DECODE(ELTON_RPC_PING_ID, struct elton_rpc_ping, in, 0,
                {/* Do nothing. */});
  return 0;
}
const static struct entry ping_entry = {
    .encode = ping_encode,
    .decode = ping_decode,
};

static int error_encode(struct packet *in, struct raw_packet **out) {
  *out = ENCODE(ELTON_RPC_ERROR_ID, struct elton_rpc_error, in, {
    enc.enc_op->u64(&enc, s->error_id);
    enc.enc_op->bytes(&enc, s->reason, strlen(s->reason));
  });
  return 0;
}
static int error_decode(struct raw_packet *in, void **out) {
  size_t reason_size;
  *out = DECODE(ELTON_RPC_ERROR_ID, struct elton_rpc_error, in, ({
                  u64 dummy;
                  dec.dec_op->u64(&dec, &dummy);
                  dec.dec_op->bytes(&dec, NULL, &reason_size);
                  reason_size + 1;
                }),
                {
                  // Initialize error.
                  s->reason = &s->__embeded_buffer;

                  dec.dec_op->u64(&dec, &s->error_id);
                  dec.dec_op->bytes(&dec, s->reason, &reason_size);
                  s->reason[reason_size] = '\0';
                });
  return 0;
}
const static struct entry error_entry = {
    .encode = error_encode,
    .decode = error_decode,
};

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
};
#define ELTON_MAX_STRUCT_ID 4

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
    kfree(raw);
  return error;
}
