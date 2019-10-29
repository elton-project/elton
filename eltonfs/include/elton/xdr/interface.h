#ifndef _ELTON_XDR_INTERFACE_H
#define _ELTON_XDR_INTERFACE_H

#include <linux/types.h>

struct xdr_encoder {
  // Buffer to write encoded data.  If NULL, the encoder only count required
  // memory size for encoding it.
  char *buffer;
  // Next write position in buffer.
  size_t pos;
  // Size of buffer.  Should be zero if buffer is NULL.
  size_t len;
  struct xdr_encoder_operations *enc_op;
};
struct xdr_decoder {
  // Buffer to read encoded data.  This field MUST NOT NULL.
  char *buffer;
  // Next read position in buffer.
  size_t pos;
  // Size of buffer.
  size_t len;
  struct xdr_decoder_operations *dec_op;
};

struct xdr_encoder_operations {
  // Encode an unsigned 8bit integer.
  int (*u8)(struct xdr_encoder *enc, u8 val);
  // Encode an unsigned 64bit integer.
  int (*u64)(struct xdr_encoder *enc, u64 val);
  // Encode bytes.
  int (*bytes)(struct xdr_encoder *enc, char *bytes, size_t len);
};
struct xdr_decoder_operations {
  // Deocde an unsigned 8bit integer.  If val is NULL, discard encoded data.
  int (*u8)(struct xdr_decoder *dec, u8 *val);
  // Decode an unsigned 64bit integer.  If val is NULL, discard encoded data.
  int (*u64)(struct xdr_decoder *dec, u64 *val);
  // Decode bytes.  If bytes is NULL, only the byte length is stored to len and
  // the discard encoded data.
  int (*bytes)(struct xdr_decoder *dec, char *bytes, size_t *len);
};

// Initialize default encoder.
// @buffer: Buffer.  If you need to the encoder only calculate required memory
//          size, it should be NULL.
// @len:    Length of buffer.  If buffer is NULL, it shoud be zero.
int default_encoder_init(struct xdr_encoder *enc, char *buff, size_t len);
// Initialize default decoder.
int default_decoder_init(struct xdr_decoder *dec, char *buff, size_t len);

#endif // _ELTON_XDR_INTERFACE_H
