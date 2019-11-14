#define ELTON_LOG_PREFIX "[rpc/server] "

#include <elton/rpc/_server.h>

#define READ_SOCK(sock, buff, size, offset)                                    \
  ({                                                                           \
    struct iovec iov = {                                                       \
        .iov_base = (buff) + (*offset),                                        \
        .iov_len = (size) - (*offset),                                         \
    };                                                                         \
    struct iov_iter iter;                                                      \
    struct kiocb kiocb;                                                        \
    ssize_t result;                                                            \
    BUG_ON((sock)->file == NULL);                                              \
    iov_iter_init(&iter, READ, &iov, 1, iov.iov_len);                          \
    init_sync_kiocb(&kiocb, (sock)->file);                                     \
    kiocb.ki_pos = 0;                                                          \
    result = (sock)->file->f_op->read_iter(&kiocb, &iter);                     \
    if (result > 0)                                                            \
      *(offset) += result;                                                     \
    result;                                                                    \
  })
#define READ_SOCK_ALL(sock, buff, size, offset)                                \
  ({                                                                           \
    ssize_t result = 0;                                                        \
    while (*offset < size) {                                                   \
      result = READ_SOCK((sock), (buff), (size), (offset));                    \
      if (result < 0)                                                          \
        /* Error */                                                            \
        break;                                                                 \
      if (result == 0)                                                         \
        /* Reached to EOF */                                                   \
        return -ELTON_RPC_EOF;                                                 \
    }                                                                          \
    result;                                                                    \
  })
#define WRITE_SOCK(sock, buff, size, offset)                                   \
  ({                                                                           \
    struct iovec iov = {                                                       \
        .iov_base = (buff) + (*offset),                                        \
        .iov_len = (size) - (*offset),                                         \
    };                                                                         \
    struct iov_iter iter;                                                      \
    struct kiocb kiocb;                                                        \
    ssize_t result;                                                            \
    BUG_ON((sock)->file == NULL);                                              \
    iov_iter_init(&iter, WRITE, &iov, 1, iov.iov_len);                         \
    init_sync_kiocb(&kiocb, (sock)->file);                                     \
    kiocb.ki_pos = 0;                                                          \
    result = (sock)->file->f_op->write_iter(&kiocb, &iter);                    \
    if (result > 0)                                                            \
      *(offset) += result;                                                     \
    result;                                                                    \
  })
#define WRITE_SOCK_ALL(sock, buff, size, offset)                               \
  ({                                                                           \
    ssize_t result = 0;                                                        \
    while (*offset < size) {                                                   \
      result = WRITE_SOCK((sock), (buff), (size), (offset));                   \
      if (result < 0)                                                          \
        break;                                                                 \
      if (result == 0) {                                                       \
        ERR("write_iter() returned zero.");                                    \
        break;                                                                 \
      }                                                                        \
    }                                                                          \
    result;                                                                    \
  })

int rpc_sock_read_packet(struct socket *sock, u64 struct_id, void **out) {
  int error = 0;
  struct raw_packet *raw = NULL;

  GOTO_IF(error_read, rpc_sock_read_raw_packet(sock, &raw));
  if (raw->struct_id != struct_id)
    GOTO_IF(error_decode, -ELTON_RPC_DIFF_TYPE);
  GOTO_IF(error_decode, elton_rpc_decode_packet(raw, out));

error_decode:
  if (raw)
    raw->free(raw);
error_read:
  return error;
}

int rpc_sock_read_raw_packet(struct socket *sock, struct raw_packet **out) {
  int error = 0;
  char buff_header[ELTON_RPC_PACKET_HEADER_SIZE];
  size_t payload_size;
  ssize_t offset = 0;
  char *buff = NULL;
  size_t buff_size;

  GOTO_IF(error_header, READ_SOCK_ALL(sock, buff_header,
                                      ELTON_RPC_PACKET_HEADER_SIZE, &offset));
  BUG_ON(ELTON_RPC_PACKET_HEADER_SIZE != offset);
  GOTO_IF(error_header,
          elton_rpc_get_raw_packet_size(buff_header, offset, &payload_size));

  buff_size = ELTON_RPC_PACKET_HEADER_SIZE + payload_size;
  buff = vmalloc(buff_size);
  if (buff == NULL)
    GOTO_IF(error_alloc_buff, -ENOMEM);

  memcpy(buff, buff_header, offset);

  GOTO_IF(error_body, READ_SOCK_ALL(sock, buff, buff_size, &offset));
  GOTO_IF(error_body, elton_rpc_build_raw_packet(out, buff, buff_size));

error_body:
  if (buff)
    vfree(buff);
error_alloc_buff:
error_header:
  return error;
}

int rpc_sock_write_raw_packet(struct socket *sock, struct raw_packet *raw) {
  int error;
  size_t offset;
  struct xdr_encoder enc;
  char buff_header[ELTON_RPC_PACKET_HEADER_SIZE];

  RETURN_IF(default_encoder_init(&enc, buff_header, sizeof(buff_header)));
  enc.enc_op->u64(&enc, raw->size);
  enc.enc_op->u64(&enc, raw->session_id);
  enc.enc_op->u8(&enc, raw->flags);
  enc.enc_op->u64(&enc, raw->struct_id);
  RETURN_IF(enc.error);

  offset = 0;
  RETURN_IF(WRITE_SOCK_ALL(sock, buff_header, sizeof(buff_header), &offset));

  offset = 0;
  RETURN_IF(WRITE_SOCK_ALL(sock, raw->data, raw->size, &offset));
  return 0;
}
