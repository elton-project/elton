#define ELTON_LOG_PREFIX "[xdr/default] "

#include <elton/assert.h>
#include <elton/xdr/bin_encoding.h>
#include <elton/xdr/interface.h>

int default_encoder_init(struct xdr_encoder *enc, char *buff, size_t len) {
  int error;
  RETURN_IF(bin_encoder_init(enc, buff, len));
  return 0;
}

int default_decoder_init(struct xdr_decoder *dec, char *buff, size_t len) {
  int error;
  RETURN_IF(bin_decoder_init(dec, buff, len));
  return 0;
}
