#ifndef _ELTON_XDR_INTERFACE_H
#define _ELTON_XDR_INTERFACE_H

#include <elton/rpc/struct.h>
#include <linux/types.h>

struct xdr_encoder {
  // Buffer to write encoded data.  If NULL, the encoder only count required
  // memory size for encoding it.
  u8 *buffer;
  // Next write position in buffer.
  size_t pos;
  // Size of buffer.  Should be zero if buffer is NULL.
  size_t len;
  // Latest error no.
  int error;
  struct xdr_encoder_operations *enc_op;
};
struct xdr_decoder {
  // Buffer to read encoded data.  This field MUST NOT NULL.
  u8 *buffer;
  // Next read position in buffer.
  size_t pos;
  // Size of buffer.
  size_t len;
  // Latest error no.
  int error;
  struct xdr_decoder_operations *dec_op;
};
struct xdr_struct_encoder {
  struct xdr_encoder *enc;
  // Number of fields.
  u8 fields;
  // Number of encoded fields.
  u8 encoded;
  // Last encoded field ID.
  u8 last_field_id;
  // Flag indicating that the encoder is closed.
  bool closed;
  struct xdr_struct_encoder_operations *op;
};
struct xdr_struct_decoder {
  struct xdr_decoder *dec;
  // Number of fields.
  u8 fields;
  // Number of decoded fields.
  u8 decoded;
  // Last decoded field ID.
  // It only using to validate input data.
  u8 last_field_id;
  // Flag indicating that the decoder is closed.
  bool closed;
  struct xdr_struct_decoder_operations *op;
};
struct xdr_map_encoder {
  struct xdr_encoder *enc;
  // Number of expected key/value pairs.
  u64 elements;
  // Number of encoded key/value pairs.
  u64 encoded;
  // Flag indicating that the decoder is closed.
  bool closed;
  struct xdr_map_encoder_operations *op;
};
struct xdr_map_decoder {
  struct xdr_decoder *dec;
  // Number of expected key/value pairs.
  u64 elements;
  // Number of decoded key/value pairs.
  u64 decoded;
  // Flag indicating that the decoder is closed.
  bool closed;
  struct xdr_map_decoder_operations *op;
};

struct xdr_encoder_operations {
  // Encode an unsigned 8bit integer.
  int (*u8)(struct xdr_encoder *enc, u8 val);
  // Encode an unsigned 64bit integer.
  int (*u64)(struct xdr_encoder *enc, u64 val);
  // Encode bytes.
  int (*bytes)(struct xdr_encoder *enc, const char *bytes, size_t len);
  // Encode timestamp.
  int (*timestamp)(struct xdr_encoder *enc, struct timestamp ts);
  // Encode struct.
  int (*struct_)(struct xdr_encoder *enc, struct xdr_struct_encoder *struct_enc,
                 u8 fields);
  // Encode map.
  int (*map)(struct xdr_encoder *enc, struct xdr_map_encoder *map_enc,
             u64 elements);
};
struct xdr_decoder_operations {
  // Deocde an unsigned 8bit integer.  If val is NULL, discard encoded data.
  int (*u8)(struct xdr_decoder *dec, u8 *val);
  // Decode an unsigned 64bit integer.  If val is NULL, discard encoded data.
  int (*u64)(struct xdr_decoder *dec, u64 *val);
  // Decode bytes.  If bytes is NULL, only the byte length is stored to len and
  // the discard encoded data.
  int (*bytes)(struct xdr_decoder *dec, char *bytes, size_t *len);
  // Decode timestamp.
  int (*timestamp)(struct xdr_decoder *dec, struct timestamp *ts);
  // Decode struct.
  int (*struct_)(struct xdr_decoder *dec,
                 struct xdr_struct_decoder *struct_dec);
  // Decode map.
  int (*map)(struct xdr_decoder *dec, struct xdr_map_decoder *map_dec);
};
struct xdr_struct_encoder_operations {
  int (*u8)(struct xdr_struct_encoder *enc, u8 field_id, u8 val);
  int (*u64)(struct xdr_struct_encoder *enc, u8 field_id, u64 val);
  int (*bytes)(struct xdr_struct_encoder *enc, u8 field_id, const char *bytes,
               size_t len);
  int (*timestamp)(struct xdr_struct_encoder *enc, u8 field_id,
                   struct timestamp ts);
  int (*map)(struct xdr_struct_encoder *enc, u8 field_id,
             struct xdr_map_encoder *map_enc, u64 elements);
  int (*close)(struct xdr_struct_encoder *enc);
  bool (*is_closed)(struct xdr_struct_encoder *enc);
};
struct xdr_struct_decoder_operations {
  int (*u8)(struct xdr_struct_decoder *dec, u8 field_id, u8 *val);
  int (*u64)(struct xdr_struct_decoder *dec, u8 field_id, u64 *val);
  int (*bytes)(struct xdr_struct_decoder *dec, u8 field_id, char *bytes,
               size_t *len);
  int (*timestamp)(struct xdr_struct_decoder *dec, u8 field_id,
                   struct timestamp *ts);
  int (*map)(struct xdr_struct_decoder *dec, u8 field_id,
             struct xdr_map_decoder *map_dec);
  int (*close)(struct xdr_struct_decoder *dec);
  bool (*is_closed)(struct xdr_struct_decoder *dec);
};
struct xdr_map_encoder_operations {
  int (*encoded_kv)(struct xdr_map_encoder *enc);
  int (*close)(struct xdr_map_encoder *enc);
  bool (*is_closed)(struct xdr_map_encoder *enc);
};
struct xdr_map_decoder_operations {
  int (*decoded_kv)(struct xdr_map_decoder *enc);
  bool (*has_next_kv)(struct xdr_map_decoder *enc);
  int (*close)(struct xdr_map_decoder *dec);
  bool (*is_closed)(struct xdr_map_decoder *dec);
};

// Initialize default encoder.
// @buffer: Buffer.  If you need to the encoder only calculate required memory
//          size, it should be NULL.
// @len:    Length of buffer.  If buffer is NULL, it shoud be zero.
int default_encoder_init(struct xdr_encoder *enc, char *buff, size_t len);
// Initialize default decoder.
int default_decoder_init(struct xdr_decoder *dec, char *buff, size_t len);

#endif // _ELTON_XDR_INTERFACE_H
