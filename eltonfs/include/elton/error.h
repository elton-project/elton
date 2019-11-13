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
// Struct field order is not valid.  Fields in the struct MUST encode/decode in
// FieldID order.
#define ELTON_XDR_INVALID_FIELD_ORDER 2004
// Decodes the field that it is not exist.
#define ELTON_XDR_NOT_FOUND_FIELD 2005
// Skipped decoding some fields.
#define ELTON_XDR_SKIP_FIELDS
// Encoded/decoded fields are not enough.  Encoder/decoder is requiring more
// fields.
#define ELTON_XDR_NOT_ENOUGH_FIELDS 2006
// Encode/decode too many fields.
#define ELTON_XDR_TOO_MANY_FIELDS 2007

#endif // _ELTON_ERROR_H
