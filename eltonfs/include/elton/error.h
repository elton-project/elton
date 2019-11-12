#ifndef _ELTON_ERROR_H
#define _ELTON_ERROR_H

#define ELTON_RPC_INVAL 1001
#define ELTON_RPC_BAD_PACKET 1002
#define ELTON_RPC_ALREADY_CLOSED 1003
// Actual struct_id does not match the expected struct_id.
#define ELTON_RPC_DIFF_TYPE 4

#define ELTON_XDR_INVAL 2001
#define ELTON_XDR_NOMEM 2002
#define ELTON_XDR_NEED_MORE_MEM 2003

#endif // _ELTON_ERROR_H
