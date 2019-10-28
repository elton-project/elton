#ifndef _ELTON_RPC_PACKET_H
#define _ELTON_RPC_PACKET_H

#include <linux/types.h>


struct packet {
    int struct_id;
    // Native data of the struct.
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
    // In many cases, this pointer points to &this.__embeded_buffer.
    char *data;
    // The function to release memory of the packet.
    void (*free)(struct raw_packet *packet);

    // Embeds encoded data at the tail of this struct.
    char __embeded_buffer;
    // WARNING: MUST NOT DEFINE ANY FIELD AFTER THE __embeded_buffer FIELD.
};


#define ELTON_RPC_PACKET_FLAG_CREATE 1
#define ELTON_RPC_PACKET_FLAG_CLOSE 2
#define ELTON_RPC_PACKET_FLAG_ERROR 3


// Encode the struct and generate raw_packet.
// The out variable sets new pointer to raw_packet.
int elton_rpc_encode_packet(struct packet *in, struct raw_packet **out);
// Decode raw_apcket.
// This out variables sets new pointer to the struct.
int elton_rpc_decode_packet(struct raw_packet *in, void **out);
// Release memory of received data.
void elton_rpc_free_decoded_data(void *data);

#endif // _ELTON_RPC_PACKET_H
