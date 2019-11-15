#define ELTON_LOG_PREFIX "[rpc/queue] "

#include <elton/assert.h>
#include <elton/rpc/queue.h>
#include <linux/sched.h>
#include <linux/slab.h>

static void elton_rpc_free(struct raw_packet *packet) { packet->free(packet); }
static inline struct elton_rpc_queue_entry *
queue_entry_init(struct raw_packet *raw) {
  struct elton_rpc_queue_entry *entry = (struct elton_rpc_queue_entry *)kmalloc(
      sizeof(struct elton_rpc_queue_entry), GFP_KERNEL);
  if (entry == NULL)
    return NULL;

  entry->raw = raw;
  return entry;
}
static void inline queue_entry_free(struct elton_rpc_queue_entry *entry) {
  kfree(entry);
}

int elton_rpc_queue_init(struct elton_rpc_queue *q) {
  spin_lock_init(&q->lock);
  init_waitqueue_head(&q->wq);
  INIT_LIST_HEAD(&q->queue);
  q->free = elton_rpc_free;
  return 0;
}

int elton_rpc_enqueue(struct elton_rpc_queue *q, struct raw_packet *raw) {
  int error = 0;
  struct elton_rpc_queue_entry *entry;

  // Allocate memory and initialize entry.
  entry = queue_entry_init(raw);
  if (entry == NULL) {
    GOTO_IF(error, -ENOMEM);
  }

  spin_lock(&q->lock);
  // Add entry to FIFO queue.
  list_add_tail(&q->queue, &entry->list_head);
  // Wake up a task.
  wake_up(&q->wq);
  spin_unlock(&q->lock);
  return 0;

error:
  return error;
}

int elton_rpc_dequeue(struct elton_rpc_queue *q, struct raw_packet **out) {
  int error = 0;
  struct elton_rpc_queue_entry *entry = NULL;

  spin_lock(&q->lock);
  // Wait until entry is enqueued.  If queue is not empty, this function returns
  // immediately.
  GOTO_IF(error_unlock, wait_event_interruptible_lock_irq(
                            q->wq, !list_empty(&q->queue), q->lock));
  // Remove entry from queue.
  BUG_ON(list_empty(&q->queue));
  entry = list_first_entry(&q->queue, struct elton_rpc_queue_entry, list_head);
  list_del(&entry->list_head);
  spin_unlock(&q->lock);

  *out = entry->raw;
  queue_entry_free(entry);
  return 0;

error_unlock:
  spin_unlock(&q->lock);
  return error;
}
