#ifndef _ELTON_ERROR_H
#define _ELTON_ERROR_H

#define ELTON_RPC_INVAL 1
#define ELTON_RPC_BAD_PACKET 2
#define ELTON_RPC_ALREADY_CLOSED 3
// Actual struct_id does not match the expected struct_id.
#define ELTON_RPC_DIFF_TYPE 4

#define ELTON_XDR_INVAL 1
#define ELTON_XDR_NOMEM 2
#define ELTON_XDR_NEED_MORE_MEM 3

#endif // _ELTON_ERROR_H
