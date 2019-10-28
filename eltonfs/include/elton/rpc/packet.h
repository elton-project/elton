#ifndef _ELTON_RPC_PACKET_H
#define _ELTON_RPC_PACKET_H

#include <linux/types.h>


struct packet {
    int struct_id;
    void *data;
};
struct raw_packet {
    // Size of data.
    size_t size;
    // Nested Session ID
    u64 session_id;
    u8 flags;
    // Struct ID
    u64 struct_id;
    // Encoded data of the struct.
    char *data;
};


#define ELTON_RPC_PACKET_FLAG_CREATE 1
#define ELTON_RPC_PACKET_FLAG_CLOSE 2
#define ELTON_RPC_PACKET_FLAG_ERROR 3


// Internal helper functions.
int elton_rpc_encode_packet(struct packet *in, struct raw_packet *out);
int elton_rpc_decode_packet(struct raw_packet *in, struct packet *out);
void elton_free_packet(struct packet *);
void elton_free_raw_packet(struct raw_packet *);


#endif // _ELTON_RPC_PACKET_H
