#include <linux/string.h>
#include "interface.h"
#include "bin_encoding.h"
#include "error.h"

int __check_encoder_status(struct xdr_encoder *enc) {
    if(
        enc->buffer == NULL ||
        enc->pos < 0 ||
        enc->len < 0 ||
        enc->len < enc->pos ||
        enc->enc_op == NULL
    ) return -ELTON_XDR_INVAL;
    return 0;
}
#define CHECK_ENCODER_STATUS(enc) \
    { \
        int err = __check_encoder_status(enc); \
        if(err) return err; \
    }

int __check_decoder_status(struct xdr_decoder *dec) {
    if(
        dec->buffer == NULL ||
        dec->pos < 0 ||
        dec->len < 0 ||
        dec->len < dec->pos ||
        dec->dec_op == NULL
    ) return -ELTON_XDR_INVAL;
    return 0;
}
#define CHECK_DECODER_STATUS(dec) \
    { \
        int err = __check_decoder_status(dec); \
        if(err) return err; \
    }

#define CHECK_READ_SIZE(xdr, size) \
    if(xdr->pos + (size) > xdr->len) return -ELTON_XDR_NEED_MORE_MEM
#define CHECK_WRITE_SIZE(xdr, size) \
    if(xdr->pos + (size) > xdr->len) return -ELTON_XDR_NOMEM


int bin_encoder_init(struct xdr_encoder *enc, char *buff, int len) {
    enc->buffer = buff;
    enc->pos = 0;
    enc->len = len;
    enc->enc_op = &bin_encoder_op;
    return 0;
}
int bin_decoder_init(struct xdr_decoder *dec, char *buff, int len) {
    dec->buffer = buff;
    dec->pos = 0;
    dec->len = len;
    dec->dec_op = &bin_decoder_op;
    return 0;
}


static int enc_u8(struct xdr_encoder *enc, u8 val) {
    CHECK_ENCODER_STATUS(enc);
    CHECK_WRITE_SIZE(enc, 1);
    enc->buffer[enc->pos++] = val;
    return 0;
}
static int enc_u64(struct xdr_encoder *enc, u64 val) {
    CHECK_ENCODER_STATUS(enc);
    CHECK_WRITE_SIZE(enc, 8);
    enc->buffer[enc->pos++] = (u8)(val>>56);
    enc->buffer[enc->pos++] = (u8)(val>>48);
    enc->buffer[enc->pos++] = (u8)(val>>40);
    enc->buffer[enc->pos++] = (u8)(val>>32);
    enc->buffer[enc->pos++] = (u8)(val>>24);
    enc->buffer[enc->pos++] = (u8)(val>>16);
    enc->buffer[enc->pos++] = (u8)(val>>8);
    enc->buffer[enc->pos++] = (u8)(val);
    return 0;
}
static int enc_bytes(struct xdr_encoder *enc, char *bytes, int len) {
    CHECK_ENCODER_STATUS(enc);
    CHECK_WRITE_SIZE(enc, len);
    memcpy(enc->buffer + enc->pos, bytes, len);
    enc->pos += len;
    return 0;
}


static int dec_u8(struct xdr_decoder *dec, u8 *val) {
    CHECK_DECODER_STATUS(dec);
    CHECK_READ_SIZE(dec, 1);
    *val = dec->buffer[dec->pos++];
    return 0;
}
static int dec_u64(struct xdr_decoder *dec, u64 *val) {
    CHECK_DECODER_STATUS(dec);
    CHECK_READ_SIZE(dec, 8);
    *val = 0;
    *val |= (u64)(dec->buffer[dec->pos++])<<56;
    *val |= (u64)(dec->buffer[dec->pos++])<<48;
    *val |= (u64)(dec->buffer[dec->pos++])<<40;
    *val |= (u64)(dec->buffer[dec->pos++])<<32;
    *val |= (u64)(dec->buffer[dec->pos++])<<24;
    *val |= (u64)(dec->buffer[dec->pos++])<<16;
    *val |= (u64)(dec->buffer[dec->pos++])<<8;
    *val |= (u64)(dec->buffer[dec->pos++]);
    return 0;
}
static int dec_bytes(struct xdr_decoder *dec, char *bytes, int *len) {
    u64 size;
    int err;
    err = dec_u64(dec, &size);
    if(err < 0) return err;

    if(len < size) {
        return -ELTON_XDR_NOMEM;
    }

    CHECK_DECODER_STATUS(dec);
    CHECK_READ_SIZE(dec, size);
    memcpy(bytes, dec->buffer + dec->pos, size);
    dec->pos += size;
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
