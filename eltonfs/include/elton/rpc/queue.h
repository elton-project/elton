#ifndef _ELTON_RPC_QUEUE_H
#define _ELTON_RPC_QUEUE_H

#include <elton/rpc/packet.h>
#include <linux/spinlock.h>
#include <linux/wait.h>

// Queue for the received packets.
struct elton_rpc_queue {
  // Global lock in the queue.
  // MUST ACQUIRE LOCK before read/write the attributes.
  struct spinlock lock;
  // Wait queue for the process waiting to receive packets.
  struct wait_queue_head wq;
  // FIFO queue of received packets.
  // If queue is empty, the entry should be NULL.
  struct elton_rpc_queue_entry *entry;
  // Free the allocated memory of raw_packet.
  // This field must not be NULL and can be accessed without locking.
  void (*free)(const struct raw_packet *packet);
};
struct elton_rpc_queue_entry {
  struct list_head list_head;
  struct raw_packet *raw;
};

// Initialize the elton_rpc_queue.
int elton_rpc_queue_init(struct elton_rpc_queue *q);
// Enqueue the received packet.
// MUST ACQUIRE LOCK before call it.
int elton_rpc_enqueue(struct elton_rpc_queue *q, struct raw_packet *in);
// Dequeue the received packet.
// The out variable will set the pointer to raw_packet.  After using it, don't
// forget free raw_packet by elton_rpc_queue.free() function. MUST ACQUIRE LOCK
// before call it.
int elton_rpc_dequeue(struct elton_rpc_queue *q, struct raw_packet **out);

#endif // _ELTON_RPC_QUEUE_H
