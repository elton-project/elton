#include <elton/assert.h>
#include <elton/rpc/packet.h>
#include <elton/rpc/struct.h>
#include <elton/xdr/interface.h>
#include <linux/bug.h>
#include <linux/vmalloc.h>

// struct_type:    Type name of the target struct.
// in:             Variable name of struct packet.
// encode_process: Statements or block that encodes a struct.
#define ENCODE(struct_type, in, encode_process)                                \
  ({                                                                           \
    struct_type *s;                                                            \
    struct xdr_encoder enc;                                                    \
    struct raw_packet *raw = NULL;                                             \
                                                                               \
    BUG_ON(in->struct_id != ELTON_RPC_SETUP1_ID);                              \
    BUG_ON(in == NULL);                                                        \
    BUG_ON(in->data == NULL);                                                  \
                                                                               \
    s = (struct elton_rpc_setup1 *)in->data;                                   \
    default_encoder_init(&enc, NULL, 0);                                       \
    do {                                                                       \
      size_t size;                                                             \
      /* The behavior is different between the first and second time.          \
       *                                                                       \
       * First time: calculate the required buffer size.                       \
       * Second time: encode data to buffer. */                                \
      encode_process;                                                          \
                                                                               \
      /* Break the loop when second time. */                                   \
      if (enc.buffer)                                                          \
        break;                                                                 \
                                                                               \
      /* Initialize raw_packet. */                                             \
      size = enc.pos;                                                          \
      raw = (struct raw_packet *)vmalloc(sizeof(struct raw_packet) + size);    \
      raw->size = size;                                                        \
      raw->struct_id = ELTON_RPC_SETUP1_ID;                                    \
      raw->free = free_raw_packet;                                             \
      raw->data = &raw->__embeded_buffer;                                      \
                                                                               \
      /* Set buffer to encoder and start the second time loop. */              \
      default_encoder_init(&enc, raw->data, raw->size);                        \
    } while (1);                                                               \
                                                                               \
    BUG_ON(raw == NULL);                                                       \
    BUG_ON(raw->size != enc.pos);                                              \
    raw;                                                                       \
  })

// struct_type:      Type name of the target struct.
// in:               Variable name of struct raw_packet.
// additional_space: The expression or statement expression to calculate
//                   additional space of struct_type.
// decode_process:   Statements or block that decodes a struct.
#define DECODE(struct_type, in, additional_space, decode_process)              \
  ({                                                                           \
    struct xdr_decoder dec;                                                    \
    size_t size;                                                               \
    struct_type *s;                                                            \
                                                                               \
    default_decoder_init(&dec, in->data, in->size);                            \
    size = additional_space;                                                   \
                                                                               \
    s = (struct_type *)kmalloc(sizeof(struct_type) + str_size + 1,             \
                               GFP_KERNEL);                                    \
    s->client_name = &s->__embeded_buffer;                                     \
                                                                               \
    default_decoder_init(&dec, in->data, in->size);                            \
    decode_process;                                                            \
    s;                                                                         \
  })

void free_raw_packet(struct raw_packet *packet) { vfree(packet); }

struct entry {
  int (*encode)(struct packet *in, struct raw_packet **out);
  int (*decode)(struct raw_packet *in, void **out);
};

static int setup1_encode(struct packet *in, struct raw_packet **out) {
  // TODO: add error handling
  // TODO: bufferがNULLでもencode出来るようにする。
  *out = ENCODE(struct elton_rpc_setup1, in, {
    enc.enc_op->bytes(&enc, s->client_name, strlen(s->client_name));
    enc.enc_op->u64(&enc, s->version_major);
    enc.enc_op->u64(&enc, s->version_minor);
    enc.enc_op->u64(&enc, s->version_revision);
  });
  return 0;
}
static int setup1_decode(struct raw_packet *in, void **out) {
  // TODO: add error handling
  // TODO: 文字列サイズだけを取得するモードを用意

  size_t str_size;
  *out = DECODE(struct elton_rpc_setup1, in, ({
                  dec.dec_op->bytes(&dec, NULL, &str_size);
                  str_size;
                }),
                {
                  dec.dec_op->bytes(&dec, s->client_name, &str_size);
                  s->client_name[str_size] = '\0';
                  dec.dec_op->u64(&dec, &s->version_major);
                  dec.dec_op->u64(&dec, &s->version_minor);
                  dec.dec_op->u64(&dec, &s->version_revision);
                });
  return 0;
}
const static struct entry setup1_entry = {
    .encode = setup1_encode,
    .decode = setup1_decode,
};
const static struct entry setup2_entry = {
    // todo
    .encode = NULL,
    .decode = NULL,
};
const static struct entry ping_entry = {
    // todo
    .encode = NULL,
    .decode = NULL,
};
const static struct entry error_entry = {
    // todo
    .encode = NULL,
    .decode = NULL,
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

  if (struct_id == 0 || struct_id > ELTON_MAX_STRUCT_ID) {
    // invalid struct id.
    // todo
  }

  *entry = look_table[struct_id];
  return 0;
}

int elton_rpc_encode_packet(struct packet *in, struct raw_packet **out) {
  const struct entry *entry;
  int error = lookup(in->struct_id, &entry);
  if (error)
    return error;

  return entry->encode(in, out);
}

int elton_rpc_decode_packet(struct raw_packet *in, void **out) {
  const struct entry *entry;
  int error = lookup(in->struct_id, &entry);
  if (error)
    return error;

  return entry->decode(in, out);
}

void elton_rpc_free_decoded_data(void *data) { kfree(data); }
