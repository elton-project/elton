#ifdef ELTONFS_UNIT_TEST
#include "interface.h"
#include "error.h"
#include <assert.h>

int test_encode_u8(void) {
    struct xdr_encoder enc;
    char buff[4] = {0, 0, 0, 99};
    int len = 3;
    int err;

    err = default_encoder_init(&enc, buff, len);
    if(err) return err;

    ASSERT_NO_ERROR(enc.enc_op->u8(&enc, 1));
    ASSERT_NO_ERROR(enc.enc_op->u8(&enc, 2));
    ASSERT_NO_ERROR(enc.enc_op->u8(&enc, 3));
    ASSERT_EQUAL_ERROR(ELTON_XDR_NOMEM, enc.enc_op->u8(&enc, 4));
    return 0;
}

#endif // ELTONFS_UNIT_TEST
