#include <linux/bug.h>
#include <linux/vmalloc.h>
#include <elton/rpc/packet.h>
#include <elton/assert.h>

struct entry {
    int (*encode)(struct packet *in, struct raw_packet **out);
    int (*decode)(struct raw_packet *in, void **out);
};

const static struct entry setup1_entry = {
    // todo
    .encode = NULL,
    .decode = NULL,
};
const static struct entry setup2_entry = {
    // todo
    .encode = NULL,
    .decode = NULL,
};
const static struct entry ping_entry = {
    // todo
    .encode = NULL,
    .decode = NULL,
};
const static struct entry error_entry = {
    // todo
    .encode = NULL,
    .decode = NULL,
};

// Lookup table from struct_id to encoder/decoder function.
const static struct entry *look_table[] = {
    // StructID 0: invalid
    NULL,
    // StructID 1: setup1
    &setup1_entry,
    // StructID 2: setup2
    &setup2_entry,
    // StructID 3: ping
    &ping_entry,
    // StructID 4: error
    &error_entry,
};
#define ELTON_MAX_STRUCT_ID 4

static int lookup(u64 struct_id, const struct entry **entry) {
    BUILD_ASSERT_EQUAL_ARRAY_SIZE(ELTON_MAX_STRUCT_ID + 1, look_table);

    if(struct_id == 0 || struct_id > ELTON_MAX_STRUCT_ID) {
        // invalid struct id.
        // todo
    }

    *entry = look_table[struct_id];
    return 0;
}

int elton_rpc_encode_packet(struct packet *in, struct raw_packet **out) {
    const struct entry *entry;
    int error = lookup(in->struct_id, &entry);
    if(error) return error;

    return entry->encode(in, out);
}

int elton_rpc_decode_packet(struct raw_packet *in, void **out) {
    const struct entry *entry;
    int error = lookup(in->struct_id, &entry);
    if(error) return error;

    return entry->decode(in, out);
}

void elton_rpc_free_decoded_data(void *data) {
    vfree(data);
}
