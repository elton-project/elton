#include "interface.h"
#include "bin_encoding.h"

int default_encoder_init(struct xdr_encoder *enc, char *buff, int len) {
    return bin_encoder_init(enc, buff, len);
}

int default_decoder_init(struct xdr_decoder *dec, char *buff, int len) {
    return bin_decoder_init(dec, buff, len);
}