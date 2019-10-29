#include <elton/xdr/bin_encoding.h>
#include <elton/xdr/interface.h>

int default_encoder_init(struct xdr_encoder *enc, char *buff, size_t len) {
  return bin_encoder_init(enc, buff, len);
}

int default_decoder_init(struct xdr_decoder *dec, char *buff, size_t len) {
  return bin_decoder_init(dec, buff, len);
}
