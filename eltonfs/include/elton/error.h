#ifndef _ELTON_ERROR_H
#define _ELTON_ERROR_H

#define ELTON_RPC_INVAL 1001
#define ELTON_RPC_BAD_PACKET 1002
#define ELTON_RPC_ALREADY_CLOSED 1003
// Actual struct_id does not match the expected struct_id.
#define ELTON_RPC_DIFF_TYPE 1004
// Reached to EOF.
#define ELTON_RPC_EOF 1005
// Received elton_rpc_error packet.
#define ELTON_RPC_ERROR_PACKET 1006

#define ELTON_XDR_INVAL 2001
#define ELTON_XDR_NOMEM 2002
#define ELTON_XDR_NEED_MORE_MEM 2003
// Struct field order is not valid.  Fields in the struct MUST encode/decode in
// FieldID order.
#define ELTON_XDR_INVALID_FIELD_ORDER 2004
// Decodes the field that it is not exist.
#define ELTON_XDR_NOT_FOUND_FIELD 2005
// Skipped decoding some fields.
#define ELTON_XDR_SKIP_FIELDS 2006
// Encoded/decoded fields are not enough.  Encoder/decoder is requiring more
// fields.
#define ELTON_XDR_NOT_ENOUGH_FIELDS 2007
// Encode/decode too many fields.  This error will occur in struct
// encoder/decoder.
#define ELTON_XDR_TOO_MANY_FIELDS 2008
// Encoder/decoder is already closed.
#define ELTON_XDR_CLOSED 2009
// Encode/decode too many elements of map. This error will occur in map
// encoder/decoder.
#define ELTON_XDR_TOO_MANY_ELEMENTS 2010

#define ELTON_CACHE_MISS 3001
#define ELTON_CACHE_LOST_LOCAL_OBJ 3002

// Max defined errno to using eltonfs.
#define ELTON_MAX_ERRNO 3002

static __always_unused void __eltonfs_errno_assert(void) {
  BUILD_BUG_ON(ELTON_MAX_ERRNO <= MAX_ERRNO);
}

#endif // _ELTON_ERROR_H
