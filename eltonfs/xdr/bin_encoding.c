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
      return err;                                                              \
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
      return err;                                                              \
    }                                                                          \
  } while (0)

#define CHECK_READ_SIZE(xdr, size)                                             \
  do {                                                                         \
    if (xdr->pos + (size) > xdr->len) {                                        \
      xdr->error = -ELTON_XDR_NEED_MORE_MEM;                                   \
      return -ELTON_XDR_NEED_MORE_MEM;                                         \
    }                                                                          \
  } while (0)

#define CHECK_WRITE_SIZE(xdr, size)                                            \
  do {                                                                         \
    if (xdr->buffer != NULL && xdr->pos + (size) > xdr->len) {                 \
      xdr->error = -ELTON_XDR_NOMEM;                                           \
      return -ELTON_XDR_NOMEM;                                                 \
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
  CHECK_ENCODER_STATUS(enc);
  CHECK_WRITE_SIZE(enc, 1);
  if (enc->buffer)
    enc->buffer[enc->pos] = val;
  enc->pos++;
  return 0;
}
static int enc_u64(struct xdr_encoder *enc, u64 val) {
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
  CHECK_ENCODER_STATUS(enc);
  CHECK_WRITE_SIZE(enc, 8 + len);
  // Write length.
  enc_u64(enc, len);
  // Write body.
  if (enc->buffer)
    memcpy(enc->buffer + enc->pos, bytes, len);
  enc->pos += len;
  return 0;
}

static int dec_u8(struct xdr_decoder *dec, u8 *val) {
  CHECK_DECODER_STATUS(dec);
  CHECK_READ_SIZE(dec, 1);
  if (val)
    *val = dec->buffer[dec->pos];
  dec->pos++;
  return 0;
}
static int dec_u64(struct xdr_decoder *dec, u64 *val) {
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
  u64 size;
  int err;
  err = dec_u64(dec, &size);
  if (err < 0)
    return err;

  if (bytes != NULL && *len < size) {
    dec->error = -ELTON_XDR_NOMEM;
    return -ELTON_XDR_NOMEM;
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

static struct xdr_encoder_operations bin_encoder_op = {
    .u8 = enc_u8,
    .u64 = enc_u64,
    .bytes = enc_bytes,
};
static struct xdr_decoder_operations bin_decoder_op = {
    .u8 = dec_u8,
    .u64 = dec_u64,
    .bytes = dec_bytes,
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
  char buff[20] = {
      1,   2,   3,   4,   5,    6,    7, 8, 0xa, 0xb,
      0xc, 0xd, 0xe, 0xf, 0x1f, 0x2f, 1, 2, 3,   4,
  };
  size_t len = 20;
  u64 val = 0;

  if (ASSERT_NO_ERROR(default_decoder_init(&dec, buff, len)))
    return;

  ASSERT_NO_ERROR(dec.dec_op->u64(&dec, &val));
  ASSERT_EQUAL_LL(0x0102030405060708LL, val);
  ASSERT_NO_ERROR(dec.dec_op->u64(&dec, &val));
  ASSERT_EQUAL_LL(0x0a0b0c0d0e0f1f2fLL, val);
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

void test_xdr_bin(void) {
  test_encode_u8();
  test_decdoe_u8();
  test_encode_u64();
  test_decode_u64();
  test_encode_bytes();
  test_decode_bytes();
}

#endif // ELTONFS_UNIT_TEST
