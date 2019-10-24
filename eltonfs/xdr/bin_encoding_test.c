#ifdef ELTONFS_UNIT_TEST
#include "interface.h"
#include <assert.h>

int test_encode_u8() {
    struct xdr_encoder enc;
    char buff[4] = {0, 0, 0, 99};
    int len = 3;
    int err;

    err = default_encoder_init(&enc, buff, len);
    if(err) return err;

    CHECK_ERROR(enc.enc_op->u8(&enc, 1));
    CHECK_ERROR(enc.enc_op->u8(&enc, 2));
    CHECK_ERROR(enc.enc_op->u8(&enc, 3));
    enc.enc_op->u8(&enc, 4);
    return 0;
}

#endif // ELTONFS_UNIT_TEST
