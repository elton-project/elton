
int bin_encoder_init(struct xdr_encoder *enc, char *buff, int len);
int bin_decoder_init(struct xdr_decoder *dec, char *buff, int len);

#ifdef ELTONFS_UNIT_TEST
void test_xdr_bin(void);
#endif // ELTONFS_UNIT_TEST
