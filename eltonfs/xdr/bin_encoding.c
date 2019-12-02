#define ELTON_LOG_PREFIX "[xdr/bin_encoding] "

#include <elton/assert.h>
#include <elton/error.h>
#include <elton/xdr/bin_encoding.h>
#include <elton/xdr/interface.h>
#include <linux/string.h>

static struct xdr_encoder_operations bin_encoder_op;
static struct xdr_decoder_operations bin_decoder_op;

int __check_encoder_status(struct xdr_encoder *enc) {
  if (!((enc->buffer == NULL && enc->len == 0) ||
        (enc->buffer != NULL && enc->pos <= enc->len)) ||
      enc->enc_op == NULL)
    return -ELTON_XDR_INVAL;
  return 0;
}
#define CHECK_ENCODER_STATUS(enc)                                              \
  do {                                                                         \
    int err = __check_encoder_status(enc);                                     \
    if (err) {                                                                 \
      enc->error = err;                                                        \
      RETURN_IF(err);                                                          \
    }                                                                          \
  } while (0)

int __check_decoder_status(struct xdr_decoder *dec) {
  if (dec->buffer == NULL || dec->len < dec->pos || dec->dec_op == NULL)
    return -ELTON_XDR_INVAL;
  return 0;
}
#define CHECK_DECODER_STATUS(dec)                                              \
  do {                                                                         \
    int err = __check_decoder_status(dec);                                     \
    if (err) {                                                                 \
      dec->error = err;                                                        \
      RETURN_IF(err);                                                          \
    }                                                                          \
  } while (0)

#define CHECK_READ_SIZE(xdr, size)                                             \
  do {                                                                         \
    if (xdr->pos + (size) > xdr->len) {                                        \
      xdr->error = -ELTON_XDR_NEED_MORE_MEM;                                   \
      RETURN_IF(-ELTON_XDR_NEED_MORE_MEM);                                     \
    }                                                                          \
  } while (0)

#define CHECK_WRITE_SIZE(xdr, size)                                            \
  do {                                                                         \
    if (xdr->buffer != NULL && xdr->pos + (size) > xdr->len) {                 \
      xdr->error = -ELTON_XDR_NOMEM;                                           \
      RETURN_IF(-ELTON_XDR_NOMEM);                                             \
    }                                                                          \
  } while (0)

int bin_encoder_init(struct xdr_encoder *enc, char *buff, size_t len) {
  enc->buffer = buff;
  enc->pos = 0;
  enc->len = len;
  enc->error = 0;
  enc->enc_op = &bin_encoder_op;
  return 0;
}
int bin_decoder_init(struct xdr_decoder *dec, char *buff, size_t len) {
  dec->buffer = buff;
  dec->pos = 0;
  dec->len = len;
  dec->error = 0;
  dec->dec_op = &bin_decoder_op;
  return 0;
}

static int enc_u8(struct xdr_encoder *enc, u8 val) {
  int error;
  CHECK_ENCODER_STATUS(enc);
  CHECK_WRITE_SIZE(enc, 1);
  if (enc->buffer)
    enc->buffer[enc->pos] = val;
  enc->pos++;
  return 0;
}
static int enc_u64(struct xdr_encoder *enc, u64 val) {
  int error;
  CHECK_ENCODER_STATUS(enc);
  CHECK_WRITE_SIZE(enc, 8);
  if (enc->buffer) {
    enc->buffer[enc->pos++] = (u8)(val >> 56);
    enc->buffer[enc->pos++] = (u8)(val >> 48);
    enc->buffer[enc->pos++] = (u8)(val >> 40);
    enc->buffer[enc->pos++] = (u8)(val >> 32);
    enc->buffer[enc->pos++] = (u8)(val >> 24);
    enc->buffer[enc->pos++] = (u8)(val >> 16);
    enc->buffer[enc->pos++] = (u8)(val >> 8);
    enc->buffer[enc->pos++] = (u8)(val);
  } else {
    enc->pos += 8;
  }
  return 0;
}
static int enc_bytes(struct xdr_encoder *enc, char *bytes, size_t len) {
  int error;
  CHECK_ENCODER_STATUS(enc);
  CHECK_WRITE_SIZE(enc, 8 + len);
  // Write length.
  RETURN_IF(enc_u64(enc, len));
  // Write body.
  if (enc->buffer)
    memcpy(enc->buffer + enc->pos, bytes, len);
  enc->pos += len;
  return 0;
}
int enc_ts(struct xdr_encoder *enc, struct timestamp ts) {
  int error;
  CHECK_ENCODER_STATUS(enc);
  CHECK_WRITE_SIZE(enc, 8 * 2);
  RETURN_IF(enc_u64(enc, ts.sec));
  RETURN_IF(enc_u64(enc, ts.nsec));
  return 0;
}

static int dec_u8(struct xdr_decoder *dec, u8 *val) {
  int error;
  CHECK_DECODER_STATUS(dec);
  CHECK_READ_SIZE(dec, 1);
  if (val)
    *val = dec->buffer[dec->pos];
  dec->pos++;
  return 0;
}
static int dec_u64(struct xdr_decoder *dec, u64 *val) {
  int error;
  CHECK_DECODER_STATUS(dec);
  CHECK_READ_SIZE(dec, 8);
  if (val) {
    *val = 0;
    *val |= (u64)(dec->buffer[dec->pos++]) << 56;
    *val |= (u64)(dec->buffer[dec->pos++]) << 48;
    *val |= (u64)(dec->buffer[dec->pos++]) << 40;
    *val |= (u64)(dec->buffer[dec->pos++]) << 32;
    *val |= (u64)(dec->buffer[dec->pos++]) << 24;
    *val |= (u64)(dec->buffer[dec->pos++]) << 16;
    *val |= (u64)(dec->buffer[dec->pos++]) << 8;
    *val |= (u64)(dec->buffer[dec->pos++]);
  } else {
    dec->pos += 8;
  }
  return 0;
}
static int dec_bytes(struct xdr_decoder *dec, char *bytes, size_t *len) {
  int error;
  u64 size;

  RETURN_IF(dec_u64(dec, &size));
  if (bytes != NULL && *len < size) {
    dec->error = -ELTON_XDR_NOMEM;
    RETURN_IF(-ELTON_XDR_NOMEM);
  }

  CHECK_DECODER_STATUS(dec);
  CHECK_READ_SIZE(dec, size);
  if (bytes)
    memcpy(bytes, dec->buffer + dec->pos, size);
  dec->pos += size;

  // Set decoded data size to len.
  *len = size;
  return 0;
}
int dec_ts(struct xdr_decoder *dec, struct timestamp *ts) {
  int error;
  CHECK_DECODER_STATUS(dec);
  CHECK_READ_SIZE(dec, 8 * 2);
  RETURN_IF(dec_u64(dec, &ts->sec));
  RETURN_IF(dec_u64(dec, &ts->nsec));
  return 0;
}

static struct xdr_struct_encoder_operations struct_encoder_op;
static struct xdr_struct_decoder_operations struct_decoder_op;

static int init_struct_encoder(struct xdr_encoder *enc,
                               struct xdr_struct_encoder *struct_enc,
                               u8 fields) {
  int error;
  BUG_ON(enc == NULL);
  BUG_ON(struct_enc == NULL);
  RETURN_IF(enc->error);

  struct_enc->enc = enc;
  struct_enc->fields = fields;
  struct_enc->encoded = 0;
  struct_enc->last_field_id = 0;
  struct_enc->closed = false;
  struct_enc->op = &struct_encoder_op;

  enc->enc_op->u8(enc, fields);
  return 0;
}
static int init_struct_decoder(struct xdr_decoder *dec,
                               struct xdr_struct_decoder *struct_dec) {
  int error;
  BUG_ON(dec == NULL);
  BUG_ON(struct_dec == NULL);
  RETURN_IF(dec->error);

  struct_dec->dec = dec;
  RETURN_IF(dec->dec_op->u8(dec, &struct_dec->fields));
  struct_dec->decoded = 0;
  struct_dec->last_field_id = 0;
  struct_dec->closed = false;
  struct_dec->op = &struct_decoder_op;
  return 0;
}

static struct xdr_map_encoder_operations map_encoder_op;
static struct xdr_map_decoder_operations map_decoder_op;

int init_map_encoder(struct xdr_encoder *enc, struct xdr_map_encoder *map_enc,
                     u64 elements) {
  int error;
  BUG_ON(enc == NULL);
  BUG_ON(map_enc == NULL);
  RETURN_IF(enc->error);

  map_enc->enc = enc;
  map_enc->elements = elements;
  map_enc->encoded = 0;
  map_enc->closed = false;
  map_enc->op = &map_encoder_op;
  return 0;
}
int init_map_decoder(struct xdr_decoder *dec, struct xdr_map_decoder *map_dec) {
  int error;
  BUG_ON(dec == NULL);
  BUG_ON(map_dec == NULL);
  RETURN_IF(dec->error);

  map_dec->dec = dec;
  RETURN_IF(dec->dec_op->u64(dec, &map_dec->elements));
  map_dec->decoded = 0;
  map_dec->closed = false;
  map_dec->op = &map_decoder_op;
  return 0;
}

static struct xdr_encoder_operations bin_encoder_op = {
    .u8 = enc_u8,
    .u64 = enc_u64,
    .bytes = enc_bytes,
    .timestamp = enc_ts,
    .struct_ = init_struct_encoder,
    .map = init_map_encoder,
};
static struct xdr_decoder_operations bin_decoder_op = {
    .u8 = dec_u8,
    .u64 = dec_u64,
    .bytes = dec_bytes,
    .timestamp = dec_ts,
    .struct_ = init_struct_decoder,
    .map = init_map_decoder,
};

#define SENC_BODY(encode_process)                                              \
  do {                                                                         \
    int error;                                                                 \
    GOTO_IF(error, senc_check(senc, field_id));                                \
    GOTO_IF(error, (encode_process));                                          \
    return 0;                                                                  \
                                                                               \
  error:                                                                       \
    senc->enc->error = error;                                                  \
    return error;                                                              \
  } while (0)
static int senc_check(struct xdr_struct_encoder *senc, u8 field_id) {
  int error;

  if (senc->closed)
    RETURN_IF(-ELTON_XDR_CLOSED);
  if (senc->fields <= senc->encoded)
    RETURN_IF(-ELTON_XDR_TOO_MANY_FIELDS);
  if (senc->last_field_id >= field_id)
    RETURN_IF(-ELTON_XDR_INVALID_FIELD_ORDER);
  RETURN_IF(senc->enc->error);

  // Write the FieldID.
  RETURN_IF(senc->enc->enc_op->u8(senc->enc, field_id));

  senc->encoded++;
  senc->last_field_id = field_id;
  return 0;
}
#define SDEC_BODY(decode_process)                                              \
  do {                                                                         \
    int error;                                                                 \
    GOTO_IF(error, sdec_check(sdec, field_id));                                \
    GOTO_IF(error, decode_process);                                            \
    return 0;                                                                  \
                                                                               \
  error:                                                                       \
    sdec->dec->error = error;                                                  \
    return error;                                                              \
  } while (0)
static int sdec_check(struct xdr_struct_decoder *sdec, u8 expected_field_id) {
  int error;
  u8 actual_field_id;

  if (sdec->closed)
    RETURN_IF(-ELTON_XDR_CLOSED);
  if (sdec->fields <= sdec->decoded)
    RETURN_IF(-ELTON_XDR_TOO_MANY_FIELDS);
  if (sdec->last_field_id >= expected_field_id)
    RETURN_IF(-ELTON_XDR_INVALID_FIELD_ORDER);
  RETURN_IF(sdec->dec->error);

  // Read the FieldID.
  RETURN_IF(sdec->dec->dec_op->u8(sdec->dec, &actual_field_id));
  if (actual_field_id <= sdec->last_field_id)
    RETURN_IF(-ELTON_XDR_INVALID_FIELD_ORDER);
  if (actual_field_id < expected_field_id)
    RETURN_IF(-ELTON_XDR_SKIP_FIELDS);
  if (actual_field_id > expected_field_id)
    RETURN_IF(-ELTON_XDR_NOT_FOUND_FIELD);
  BUG_ON(actual_field_id != expected_field_id);

  sdec->decoded++;
  sdec->last_field_id = expected_field_id;
  return 0;
}
static int senc_u8(struct xdr_struct_encoder *senc, u8 field_id, u8 val) {
  SENC_BODY(senc->enc->enc_op->u8(senc->enc, val));
}
static int sdec_u8(struct xdr_struct_decoder *sdec, u8 field_id, u8 *val) {
  SDEC_BODY(sdec->dec->dec_op->u8(sdec->dec, val));
}
static int senc_u64(struct xdr_struct_encoder *senc, u8 field_id, u64 val) {
  SENC_BODY(senc->enc->enc_op->u64(senc->enc, val));
}
static int sdec_u64(struct xdr_struct_decoder *sdec, u8 field_id, u64 *val) {
  SDEC_BODY(sdec->dec->dec_op->u64(sdec->dec, val));
}
static int senc_bytes(struct xdr_struct_encoder *senc, u8 field_id, char *bytes,
                      size_t len) {
  SENC_BODY(senc->enc->enc_op->bytes(senc->enc, bytes, len));
}
static int sdec_bytes(struct xdr_struct_decoder *sdec, u8 field_id, char *bytes,
                      size_t *len) {
  SDEC_BODY(sdec->dec->dec_op->bytes(sdec->dec, bytes, len));
}
static int senc_close(struct xdr_struct_encoder *senc) {
  int error;
  if (senc->closed)
    return senc->enc->error;
  if (senc->encoded < senc->fields)
    GOTO_IF(error, -ELTON_XDR_NOT_ENOUGH_FIELDS);
  if (senc->encoded != senc->fields)
    GOTO_IF(error, -ELTON_XDR_NOT_ENOUGH_FIELDS);
  return senc->enc->error;

error:
  senc->enc->error = error;
  return error;
}
int senc_ts(struct xdr_struct_encoder *senc, u8 field_id, struct timestamp ts) {
  SENC_BODY(senc->enc->enc_op->timestamp(senc->enc, ts));
}
int sdec_ts(struct xdr_struct_decoder *sdec, u8 field_id,
            struct timestamp *ts) {
  SDEC_BODY(sdec->dec->dec_op->timestamp(sdec->dec, ts));
}
static int senc_map(struct xdr_struct_encoder *senc, u8 field_id,
                    struct xdr_map_encoder *map_enc, u64 elements) {
  SENC_BODY(senc->enc->enc_op->map(senc->enc, map_enc, elements));
}
static int sdec_map(struct xdr_struct_decoder *sdec, u8 field_id,
                    struct xdr_map_decoder *map_dec) {
  SDEC_BODY(sdec->dec->dec_op->map(sdec->dec, map_dec));
}
static int sdec_close(struct xdr_struct_decoder *sdec) {
  int error;
  if (sdec->closed)
    return sdec->dec->error;
  if (sdec->decoded < sdec->fields)
    GOTO_IF(error, -ELTON_XDR_NOT_ENOUGH_FIELDS);
  if (sdec->decoded != sdec->fields)
    GOTO_IF(error, -ELTON_XDR_NOT_ENOUGH_FIELDS);
  return sdec->dec->error;

error:
  sdec->dec->error = error;
  return error;
}
bool senc_is_closed(struct xdr_struct_encoder *enc) { return enc->closed; }
bool sdec_is_closed(struct xdr_struct_decoder *dec) { return dec->closed; }

static struct xdr_struct_encoder_operations struct_encoder_op = {
    .u8 = senc_u8,
    .u64 = senc_u64,
    .bytes = senc_bytes,
    .timestamp = senc_ts,
    .map = senc_map,
    .close = senc_close,
    .is_closed = senc_is_closed,
};
static struct xdr_struct_decoder_operations struct_decoder_op = {
    .u8 = sdec_u8,
    .u64 = sdec_u64,
    .bytes = sdec_bytes,
    .timestamp = sdec_ts,
    .map = sdec_map,
    .close = sdec_close,
    .is_closed = sdec_is_closed,
};

// todo: impl
int menc_u8(struct xdr_map_encoder *enc, u8 val);
int mdec_u8(struct xdr_map_decoder *dec, u8 *val);
int menc_u64(struct xdr_map_encoder *enc, u64 val);
int mdec_u64(struct xdr_map_decoder *dec, u64 *val);
int menc_bytes(struct xdr_map_encoder *enc, char *bytes, size_t len);
int mdec_bytes(struct xdr_map_decoder *dec, char *bytes, size_t *len);
int menc_timestamp(struct xdr_map_encoder *enc, struct timestamp ts);
int mdec_timestamp(struct xdr_map_decoder *dec, struct timestamp *ts);
int menc_struct_(struct xdr_map_encoder *enc,
                 struct xdr_struct_encoder *struct_enc, u8 fields);
int mdec_struct_(struct xdr_map_decoder *dec,
                 struct xdr_struct_decoder *struct_dec);
int menc_close(struct xdr_map_encoder *enc);
int mdec_close(struct xdr_map_decoder *dec);
bool menc_is_closed(struct xdr_map_encoder *enc);
bool mdec_is_closed(struct xdr_map_decoder *dec);

static struct xdr_map_encoder_operations map_encoder_op = {
    .u8 = menc_u8,
    .u64 = menc_u64,
    .bytes = menc_bytes,
    .timestamp = menc_timestamp,
    .struct_ = menc_struct_,
    .close = menc_close,
    .is_closed = menc_is_closed,
};
static struct xdr_map_decoder_operations map_decoder_op = {
    .u8 = mdec_u8,
    .u64 = mdec_u64,
    .bytes = mdec_bytes,
    .timestamp = mdec_timestamp,
    .struct_ = mdec_struct_,
    .close = mdec_close,
    .is_closed = mdec_is_closed,
};

#ifdef ELTONFS_UNIT_TEST

static void test_encode_u8(void) {
  struct xdr_encoder enc;
  char buff[4] = {0, 0, 0, 99};
  char expected[] = {1, 2, 3};
  size_t len = 3;

  if (ASSERT_NO_ERROR(default_encoder_init(&enc, buff, len)))
    return;

  ASSERT_NO_ERROR(enc.enc_op->u8(&enc, 1));
  ASSERT_NO_ERROR(enc.enc_op->u8(&enc, 2));
  ASSERT_NO_ERROR(enc.enc_op->u8(&enc, 3));
  ASSERT_EQUAL_ERROR(-ELTON_XDR_NOMEM, enc.enc_op->u8(&enc, 4));
  ASSERT_EQUAL_BYTES(expected, buff, sizeof(expected));

  // Check out-of-bounds writing.
  ASSERT_EQUAL_INT(99, buff[3]);

  // Test for discard mode.
  if (ASSERT_NO_ERROR(default_encoder_init(&enc, NULL, 0)))
    return;
  ASSERT_NO_ERROR(enc.enc_op->u8(&enc, 1));
  ASSERT_NO_ERROR(enc.enc_op->u8(&enc, 2));
}

static void test_decdoe_u8(void) {
  struct xdr_decoder dec;
  char buff[4] = {1, 2, 3, 99};
  size_t len = 3;
  u8 val = 0;

  if (ASSERT_NO_ERROR(default_decoder_init(&dec, buff, len)))
    return;

  ASSERT_NO_ERROR(dec.dec_op->u8(&dec, &val));
  ASSERT_EQUAL_ERROR(1, val);
  ASSERT_NO_ERROR(dec.dec_op->u8(&dec, &val));
  ASSERT_EQUAL_ERROR(2, val);
  ASSERT_NO_ERROR(dec.dec_op->u8(&dec, &val));
  ASSERT_EQUAL_ERROR(3, val);
  ASSERT_EQUAL_ERROR(-ELTON_XDR_NEED_MORE_MEM, dec.dec_op->u8(&dec, &val));

  // Test for discard mode.
  if (ASSERT_NO_ERROR(default_decoder_init(&dec, buff, len)))
    return;
  ASSERT_NO_ERROR(dec.dec_op->u8(&dec, NULL));
  ASSERT_NO_ERROR(dec.dec_op->u8(&dec, NULL));
}

static void test_encode_u64(void) {
  struct xdr_encoder enc;
  char buff[8 * 3];
  size_t len = 8 * 2 + 4;
  const char expected1[] = {1, 2, 3, 4, 5, 6, 7, 8};
  const char expected2[] = {0xa, 0xb, 0xc, 0xd, 0xe, 0xf, 0x1f, 0x2f};
  const char expected3[] = {201, 202, 203, 204, 205, 206, 207, 208};

  if (ASSERT_NO_ERROR(default_encoder_init(&enc, buff, len)))
    return;

  ASSERT_NO_ERROR(enc.enc_op->u64(&enc, 0x0102030405060708));
  ASSERT_EQUAL_BYTES(expected1, buff, 8);
  ASSERT_NO_ERROR(enc.enc_op->u64(&enc, 0x0a0b0c0d0e0f1f2f));
  ASSERT_EQUAL_BYTES(expected2, buff + 8, 8);

  // Check out-of-bounds writing.
  memcpy(buff + (8 * 2), expected3, 8);
  ASSERT_EQUAL_ERROR(-ELTON_XDR_NOMEM, enc.enc_op->u64(&enc, 0x123));
  ASSERT_EQUAL_BYTES(expected3, buff + (8 * 2), 8);

  // Test for discard mode.
  if (ASSERT_NO_ERROR(default_encoder_init(&enc, NULL, 0)))
    return;
  ASSERT_NO_ERROR(enc.enc_op->u64(&enc, 123));
  ASSERT_NO_ERROR(enc.enc_op->u64(&enc, 456));
}
static void test_decode_u64(void) {
  struct xdr_decoder dec;
  char buff[] = {
      1,   2,   3,   4,   5,   6,   7,    8,    // First value
      0xa, 0xb, 0xc, 0xd, 0xe, 0xf, 0x1f, 0x2f, // Second value
      0,   0,   0,   0,   0,   0,   0,    128,  // Third value
      1,   2,   3,   4,                         // Padding
  };
  const size_t len = sizeof(buff);
  u64 val = 0;

  if (ASSERT_NO_ERROR(default_decoder_init(&dec, buff, len)))
    return;

  ASSERT_NO_ERROR(dec.dec_op->u64(&dec, &val));
  ASSERT_EQUAL_LL(0x0102030405060708LL, val);
  ASSERT_NO_ERROR(dec.dec_op->u64(&dec, &val));
  ASSERT_EQUAL_LL(0x0a0b0c0d0e0f1f2fLL, val);
  ASSERT_NO_ERROR(dec.dec_op->u64(&dec, &val));
  ASSERT_EQUAL_LL(128LL, val);
  ASSERT_EQUAL_ERROR(-ELTON_XDR_NEED_MORE_MEM, dec.dec_op->u64(&dec, &val));

  // Test for discard mode.
  if (ASSERT_NO_ERROR(default_decoder_init(&dec, buff, len)))
    return;
  ASSERT_NO_ERROR(dec.dec_op->u64(&dec, NULL));
  ASSERT_NO_ERROR(dec.dec_op->u64(&dec, NULL));
}
static void test_encode_bytes(void) {
  struct xdr_encoder enc;
  char buff[37];
  size_t len = 30;
  char *data1 = "hello";
  char *data2 = "world!!";
  char *data3 = "long-long-data";
  char expected12[] = {
      0,   0,   0,   0,   0,   0,   0,   5, // length
      'h', 'e', 'l', 'l', 'o',              // data1
      0,   0,   0,   0,   0,   0,   0,   7, // length
      'w', 'o', 'r', 'l', 'd', '!', '!',    // data2
      0,   0,   0,   0,   0,   0,   0,   0,
      0, // padding for detect out-of-bounds writing.
  };
  BUILD_BUG_ON_MSG(sizeof(buff) != sizeof(expected12),
                   "mismatch data size of buff and expected12");

  if (ASSERT_NO_ERROR(default_encoder_init(&enc, buff, len)))
    return;

  memset(buff, 0, sizeof(buff));
  ASSERT_NO_ERROR(enc.enc_op->bytes(&enc, data1, strlen(data1)));
  ASSERT_NO_ERROR(enc.enc_op->bytes(&enc, data2, strlen(data2)));
  ASSERT_EQUAL_BYTES(expected12, buff, sizeof(buff));

  // Check out-of-bounds writing.
  ASSERT_EQUAL_ERROR(-ELTON_XDR_NOMEM,
                     enc.enc_op->bytes(&enc, data3, strlen(data3)));
  ASSERT_EQUAL_BYTES(expected12, buff, sizeof(buff));

  // Test for discard mode.
  if (ASSERT_NO_ERROR(default_encoder_init(&enc, NULL, 0)))
    return;
  ASSERT_NO_ERROR(enc.enc_op->bytes(&enc, data1, strlen(data1)));
  ASSERT_NO_ERROR(enc.enc_op->bytes(&enc, data2, strlen(data2)));
}
static void test_decode_bytes(void) {
  struct xdr_decoder dec;
  char buff[] = {
      0,   0,   0,   0,   0,   0,   0,   5, // length
      'h', 'e', 'l', 'l', 'o',              // data1
      0,   0,   0,   0,   0,   0,   0,   7, // length
      'w', 'o', 'r', 'l', 'd', '!', '!',    // data2
      0,   0,   0,   0,   0,   0,   0,   5, // length
      'a', 'b',                             // The partial data
  };
  char read_buff[10];
  size_t read_size;
  char *expected1 = "hello";
  char *expected2 = "world!!";

  if (ASSERT_NO_ERROR(default_decoder_init(&dec, buff, sizeof(buff))))
    return;

  read_size = sizeof(read_buff);
  ASSERT_NO_ERROR(dec.dec_op->bytes(&dec, read_buff, &read_size));
  ASSERT_EQUAL_SIZE_T(strlen(expected1), read_size);
  ASSERT_EQUAL_BYTES(expected1, read_buff, strlen(expected1));

  read_size = sizeof(read_buff);
  ASSERT_NO_ERROR(dec.dec_op->bytes(&dec, read_buff, &read_size));
  ASSERT_EQUAL_SIZE_T(strlen(expected2), read_size);
  ASSERT_EQUAL_BYTES(expected2, read_buff, strlen(expected2));

  read_size = sizeof(read_buff);
  ASSERT_EQUAL_ERROR(-ELTON_XDR_NEED_MORE_MEM,
                     dec.dec_op->bytes(&dec, read_buff, &read_size));

  // Test for discard mode.
  if (ASSERT_NO_ERROR(default_decoder_init(&dec, buff, sizeof(buff))))
    return;
  read_size = 0;
  ASSERT_NO_ERROR(dec.dec_op->bytes(&dec, NULL, &read_size));
  ASSERT_EQUAL_SIZE_T(strlen(expected1), read_size);
  read_size = 0;
  ASSERT_NO_ERROR(dec.dec_op->bytes(&dec, NULL, &read_size));
  ASSERT_EQUAL_SIZE_T(strlen(expected2), read_size);
}
static void test_encode_struct(void) {
  char buff[100];
  struct xdr_encoder enc;
  struct xdr_struct_encoder se;
  u8 val8 = 8;
  u64 val64 = 64;

  // Test for normal case.
  {
    char expected[] = {
        3,                          // Number of fields.
        1, 8,                       // Field 1
        2, 0, 0, 0, 0, 0, 0, 0, 64, // Field 2

        // Field 3
        3,                       // Field ID
        0, 0, 0, 0, 0, 0, 0, 5,  // Length of bytes.
        'h', 'e', 'l', 'l', 'o', // Body of bytes.
    };
    if (ASSERT_NO_ERROR(default_encoder_init(&enc, buff, sizeof(buff))))
      return;
    if (ASSERT_NO_ERROR(enc.enc_op->struct_(&enc, &se, 3)))
      return;
    ASSERT_NO_ERROR(se.op->u8(&se, 1, val8));
    ASSERT_NO_ERROR(se.op->u64(&se, 2, val64));
    ASSERT_NO_ERROR(se.op->bytes(&se, 3, "hello", 5));
    ASSERT_NO_ERROR(se.op->close(&se));
    ASSERT_NO_ERROR(enc.error);
    ASSERT_EQUAL_SIZE_T(sizeof(expected), enc.pos);
    ASSERT_EQUAL_BYTES(expected, buff, enc.pos);
  }

  // Test for error handling of invalid field order.
  {
    if (ASSERT_NO_ERROR(default_encoder_init(&enc, buff, sizeof(buff))))
      return;
    if (ASSERT_NO_ERROR(enc.enc_op->struct_(&enc, &se, 2)))
      return;
    ASSERT_NO_ERROR(se.op->u64(&se, 2, val64));
    ASSERT_EQUAL_ERROR(-ELTON_XDR_INVALID_FIELD_ORDER, se.op->u8(&se, 1, val8));
    ASSERT_EQUAL_ERROR(-ELTON_XDR_INVALID_FIELD_ORDER, se.op->u8(&se, 2, val8));
  }

  // Test for error handling of not enough fields are written.
  {
    if (ASSERT_NO_ERROR(default_encoder_init(&enc, buff, sizeof(buff))))
      return;
    if (ASSERT_NO_ERROR(enc.enc_op->struct_(&enc, &se, 2)))
      return;
    ASSERT_NO_ERROR(se.op->u8(&se, 1, val8));
    ASSERT_EQUAL_ERROR(-ELTON_XDR_NOT_ENOUGH_FIELDS, se.op->close(&se));
  }

  // Test for error handling of too many fields.
  if (ASSERT_NO_ERROR(default_encoder_init(&enc, buff, sizeof(buff))))
    return;
  if (ASSERT_NO_ERROR(enc.enc_op->struct_(&enc, &se, 2)))
    return;
  ASSERT_NO_ERROR(se.op->u8(&se, 1, val8));
  ASSERT_NO_ERROR(se.op->u8(&se, 2, val8));
  ASSERT_EQUAL_ERROR(-ELTON_XDR_TOO_MANY_FIELDS, se.op->u64(&se, 3, val64));
}
static void test_decode_struct(void) {
  struct xdr_decoder dec;
  struct xdr_struct_decoder sd;
  u8 val8;
  u64 val64;

  // Test for normal case.
  {
    char buff[] = {
        3,                          // Number of fields.
        1, 8,                       // Field 1
        2, 0, 0, 0, 0, 0, 0, 0, 64, // Field 2

        // Field 3
        3,                       // Field ID
        0, 0, 0, 0, 0, 0, 0, 5,  // Length of bytes.
        'h', 'e', 'l', 'l', 'o', // Body of bytes.
    };
    char bytes_buff[10];
    size_t bytes_size;
    if (ASSERT_NO_ERROR(default_decoder_init(&dec, buff, sizeof(buff))))
      return;
    if (ASSERT_NO_ERROR(dec.dec_op->struct_(&dec, &sd)))
      return;
    ASSERT_NO_ERROR(sd.op->u8(&sd, 1, &val8));
    ASSERT_EQUAL_INT(8, val8);
    ASSERT_NO_ERROR(sd.op->u64(&sd, 2, &val64));
    ASSERT_EQUAL_INT(64, val64);
    bytes_size = sizeof(bytes_buff);
    ASSERT_NO_ERROR(sd.op->bytes(&sd, 3, bytes_buff, &bytes_size));
    ASSERT_EQUAL_SIZE_T((size_t)5, bytes_size);
    ASSERT_EQUAL_BYTES("hello", bytes_buff, 5);
    ASSERT_NO_ERROR(sd.op->close(&sd));
    ASSERT_NO_ERROR(dec.error);
    ASSERT_EQUAL_SIZE_T(sizeof(buff), dec.pos);
  }

  // Test for error handling of invalid field order.
  {
    char buff[] = {
        2,    // Number of fields
        3, 8, // Field 3
        4, 1, // Field 4
    };
    if (ASSERT_NO_ERROR(default_decoder_init(&dec, buff, sizeof(buff))))
      return;
    if (ASSERT_NO_ERROR(dec.dec_op->struct_(&dec, &sd)))
      return;
    ASSERT_NO_ERROR(sd.op->u8(&sd, 3, &val8));
    ASSERT_EQUAL_ERROR(-ELTON_XDR_INVALID_FIELD_ORDER,
                       sd.op->u8(&sd, 1, &val8));
  }

  // Test for error handling of not enough fields are readed.
  {
    char buff[] = {
        2,    // Number of fields
        1, 8, // Field 1
        2, 4, // Field 2
    };
    if (ASSERT_NO_ERROR(default_decoder_init(&dec, buff, sizeof(buff))))
      return;
    if (ASSERT_NO_ERROR(dec.dec_op->struct_(&dec, &sd)))
      return;
    ASSERT_NO_ERROR(sd.op->u8(&sd, 1, &val8));
    ASSERT_EQUAL_ERROR(-ELTON_XDR_NOT_ENOUGH_FIELDS, sd.op->close(&sd));
  }

  // Test for error handling of too many fields are readed.
  {
    char buff[] = {
        1,    // Number of fields
        1, 8, // Field 1
    };
    if (ASSERT_NO_ERROR(default_decoder_init(&dec, buff, sizeof(buff))))
      return;
    if (ASSERT_NO_ERROR(dec.dec_op->struct_(&dec, &sd)))
      return;
    ASSERT_NO_ERROR(sd.op->u8(&sd, 1, &val8));
    ASSERT_EQUAL_ERROR(-ELTON_XDR_TOO_MANY_FIELDS, sd.op->u8(&sd, 2, &val8));
  }
}

void test_xdr_bin(void) {
  test_encode_u8();
  test_decdoe_u8();
  test_encode_u64();
  test_decode_u64();
  test_encode_bytes();
  test_decode_bytes();
  test_encode_struct();
  test_decode_struct();
}

#endif // ELTONFS_UNIT_TEST
