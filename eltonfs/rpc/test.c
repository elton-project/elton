#include <elton/rpc/server.h>
#include <elton/rpc/test.h>
#include <elton/assert.h>
#include <elton/rpc/error.h>
#include <elton/rpc/struct.h>
#include <elton/rpc/packet.h>
#include <elton/rpc/queue.h>


void test_encode_packet(void) {
    struct raw_packet out;
    struct elton_rpc_ping ping;
    struct packet p1 = {
        .struct_id = 0,
        .data = &ping,
    };
    struct packet p2 = {
        .struct_id = 3,
        .data = NULL,
    };
    struct packet p3 = {
        .struct_id = 3,
        .data = &ping,
    };

    // Invalid struct_id.
    ASSERT_EQUAL_ERROR(-ELTON_RPC_INVAL, elton_rpc_encode_packet(&p1, &out));
    elton_free_raw_packet(&out);
    // Invalid data pointer.
    ASSERT_EQUAL_ERROR(-ELTON_RPC_INVAL, elton_rpc_encode_packet(&p2, &out));
    elton_free_raw_packet(&out);
    // Valid data.
    ASSERT_NO_ERROR(elton_rpc_encode_packet(&p3, &out));
    elton_free_raw_packet(&out);
}
void test_decode_packet(void) {
    struct packet out;
    char ping[] = {
        0, 0, 0, 0, 0, 0, 0, 0,  // Size of this packet.
        0, 0, 0, 0, 0, 0, 0, 1,  // Nested Session ID
        ELTON_RPC_PACKET_FLAG_CREATE,  // Packet Flags
        0, 0, 0, 0, 0, 0, 0, 3,  // Struct ID (ping)
        // data (ping)
    };
    char bad_ping[] = {};
    struct raw_packet p1 = {
        .struct_id = 0,
        .data = ping,
        .size = sizeof(ping),
    };
    struct raw_packet p2 = {
        .struct_id = 3,
        .data = NULL,
        .size = sizeof(ping),
    };
    struct raw_packet p3 = {
        .struct_id = 3,
        .data = ping,
        .size = 0
    };
    struct raw_packet p4 = {
        .struct_id = 3,
        .data = bad_ping,
        .size = sizeof(bad_ping),
    };
    struct raw_packet p5 = {
        .struct_id = 3,
        .data = ping,
        .size = sizeof(ping),
    };

    // Invalid struct_id.
    ASSERT_EQUAL_ERROR(-ELTON_RPC_INVAL, elton_rpc_decode_packet(&p1, &out));
    elton_free_packet(&out);
    // Invalid data pointer.
    ASSERT_EQUAL_ERROR(-ELTON_RPC_INVAL, elton_rpc_decode_packet(&p2, &out));
    elton_free_packet(&out);
    // Invalid data_size.
    ASSERT_EQUAL_ERROR(-ELTON_RPC_INVAL, elton_rpc_decode_packet(&p3, &out));
    // Bad packet.
    ASSERT_EQUAL_ERROR(-ELTON_RPC_BAD_PACKET, elton_rpc_decode_packet(&p4, &out));
    // Valid data.
    ASSERT_NO_ERROR(elton_rpc_decode_packet(&p5, &out));
}
void test_packet_queue(void) {
    struct elton_rpc_queue q;
    struct raw_packet *in, out;

    elton_rpc_queue_init(&q);
    elton_rpc_enqueue(&q, in); // todo
    elton_rpc_dequeue(&q, &out); // todo
}

static void test_server(void) {
    test_encode_packet();
    test_packet_queue();
}

void test_rpc(void) {
    test_server();
}
