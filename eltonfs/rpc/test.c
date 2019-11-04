#include <elton/assert.h>
#include <elton/compiler_attributes.h>
#include <elton/rpc/error.h>
#include <elton/rpc/packet.h>
#include <elton/rpc/queue.h>
#include <elton/rpc/server.h>
#include <elton/rpc/struct.h>
#include <elton/rpc/test.h>
#include <linux/bug.h>
#include <linux/vmalloc.h>

static char encoded_ping[] = {
    // Size of this packet
    0, 0, 0, 0, 0, 0, 0, 0,
    // Nested Session ID
    0, 0, 0, 0, 0, 0, 0, 1,
    // Packet Flags
    ELTON_RPC_PACKET_FLAG_CREATE,
    // Struct ID (ping)
    0, 0, 0, 0, 0, 0, 0, 3,
    // data (ping)
};
static char bad_encoded_ping[] = {};

static void __unused ___elton_rpc_test___detect_build_bug(void) {
  BUILD_BUG_ON_MSG((8 * 3 + 1) != sizeof(encoded_ping), "unexpected size");
  BUILD_BUG_ON_MSG(0 != sizeof(bad_encoded_ping), "unexpected size");
}

void vfree_raw_packet(struct raw_packet *packet) { vfree(packet); }

void test_encode_packet(void) {
  struct raw_packet *out;
  struct elton_rpc_ping ping;
  struct packet p1 = {
      .struct_id = 3,
      .data = &ping,
  };

  // Valid data.
  ASSERT_NO_ERROR(elton_rpc_encode_packet(&p1, &out));
  out->free(out);
  // Invalid data.
  // MUST panic while encoding.
}
void test_decode_packet(void) {
  struct ping *out;
  struct raw_packet p1 = {
      .struct_id = 0,
      .data = encoded_ping,
      .size = sizeof(encoded_ping),
  };
  struct raw_packet p2 = {
      .struct_id = 3,
      .data = NULL,
      .size = sizeof(encoded_ping),
  };
  struct raw_packet p3 = {.struct_id = 3, .data = encoded_ping, .size = 0};
  struct raw_packet p4 = {
      .struct_id = 3,
      .data = bad_encoded_ping,
      .size = sizeof(bad_encoded_ping),
  };
  struct raw_packet p5 = {
      .struct_id = 3,
      .data = encoded_ping,
      .size = sizeof(encoded_ping),
  };

  // Invalid struct_id.
  ASSERT_EQUAL_ERROR(-ELTON_RPC_INVAL,
                     elton_rpc_decode_packet(&p1, (void **)&out));
  // Invalid data pointer.
  ASSERT_EQUAL_ERROR(-ELTON_RPC_INVAL,
                     elton_rpc_decode_packet(&p2, (void **)&out));
  // Invalid data_size.
  ASSERT_EQUAL_ERROR(-ELTON_RPC_INVAL,
                     elton_rpc_decode_packet(&p3, (void **)&out));
  // Bad packet.
  ASSERT_EQUAL_ERROR(-ELTON_RPC_BAD_PACKET,
                     elton_rpc_decode_packet(&p4, (void **)&out));
  // Valid data.
  ASSERT_NO_ERROR(elton_rpc_decode_packet(&p5, (void **)&out));
  elton_rpc_free_decoded_data(out);
}
void test_packet_queue(void) {
  struct elton_rpc_queue *q;
  struct raw_packet *in, *out;

  // Intialize the in.
  in = vzalloc(sizeof(struct raw_packet));
  in->struct_id = 3;
  in->data = encoded_ping;
  in->size = sizeof(encoded_ping);
  in->free = vfree_raw_packet;

  // Initialize the q.
  q = vmalloc(sizeof(struct elton_rpc_queue));
  if (ASSERT_NOT_NULL(q))
    return; // Memory allocation error.
  elton_rpc_queue_init(q);

  out = NULL;
  ASSERT_NO_ERROR(elton_rpc_enqueue(q, in));
  ASSERT_NO_ERROR(elton_rpc_dequeue(q, &out));
  if (out)
    q->free(out);
}

static void test_server(void) {
  test_encode_packet();
  test_packet_queue();
}

void test_rpc(void) { test_server(); }
