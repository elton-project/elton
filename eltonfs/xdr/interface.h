#include <linux/types.h>

struct xdr_encoder {
    char *buffer;
    int pos;
    int len;
    struct xdr_encoder_operations *enc_op;
};
struct xdr_decoder {
    char *buffer;
    int pos;
    int len;
    struct xdr_decoder_operations *dec_op;
};

struct xdr_encoder_operations {
    int (*u8)(struct xdr_encoder *enc, u8 val);
    int (*u64)(struct xdr_encoder *enc, u64 val);
    int (*bytes)(struct xdr_encoder *enc, char *bytes, int len);
};
struct xdr_decoder_operations {
    int (*u8)(struct xdr_decoder *dec, u8 *val);
    int (*u64)(struct xdr_decoder *dec, u64 *val);
    int (*bytes)(struct xdr_decoder *dec, char *bytes, int *len);
};

int default_encoder_init(struct xdr_encoder *enc, char *buff, int len);
int default_decoder_init(struct xdr_decoder *dec, char *buff, int len);
