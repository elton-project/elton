#ifndef _ELTON_XDR_BIN_ENCODING_H
#define _ELTON_XDR_BIN_ENCODING_H

#include <elton/xdr/interface.h>
#include <linux/types.h>

// Initialize simple binary encoer.  See help of default_encoder_init() about
// args.
int bin_encoder_init(struct xdr_encoder *enc, char *buff, size_t len);
// Initialize simple binary decoder.  See help of default_decoder_init() about
// args.
int bin_decoder_init(struct xdr_decoder *dec, char *buff, size_t len);

#ifdef ELTONFS_UNIT_TEST
void test_xdr_bin(void);
#endif // ELTONFS_UNIT_TEST

#endif // _ELTON_XDR_BIN_ENCODING_H
