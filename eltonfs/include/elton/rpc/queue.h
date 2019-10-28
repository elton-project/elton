#ifndef _ELTON_RPC_QUEUE_H
#define _ELTON_RPC_QUEUE_H

#include <elton/rpc/packet.h>

struct elton_rpc_queue {
    int head, tail;
    // Circular buffer.
    // Use CIRC_* macros defined in linux/circ_buf.h
    struct raw_packet *buffer[1<<10];
    // todo: https://debimate.jp/2019/07/07/linux-kernel-mutex-api%E3%81%AB%E3%82%88%E3%82%8B%E3%83%AD%E3%83%83%E3%82%AF%E6%8E%92%E4%BB%96%E6%96%B9%E6%B3%95/
};

int elton_rpc_queue_init(struct elton_rpc_queue *q);
int elton_rpc_enqueue(struct elton_rpc_queue *q, struct raw_packet *in);
int elton_rpc_dequeue(struct elton_rpc_queue *q, struct raw_packet *out);


#endif // _ELTON_RPC_QUEUE_H
