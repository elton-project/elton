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

  INIT_LIST_HEAD(&entry->list_head);
  entry->raw = raw;
  return entry;
}
static void inline queue_entry_free(struct elton_rpc_queue_entry *entry) {
  kfree(entry);
}

int elton_rpc_queue_init(struct elton_rpc_queue *q) {
  spin_lock_init(&q->lock);
  init_waitqueue_head(&q->wq);
  q->entry = NULL;
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
  if (q->entry) {
    list_add_tail_rcu(&entry->list_head, &q->entry->list_head);
  } else {
    INIT_LIST_HEAD(&entry->list_head);
    q->entry = entry;
  }
  // Wake up a task.
  wake_up(&q->wq);
  spin_unlock(&q->lock);
  return 0;

error:
  return error;
}

int elton_rpc_dequeue(struct elton_rpc_queue *q, struct raw_packet **out) {
  int error = 0;

  spin_lock(&q->lock);
  // Wait until entry is enqueued.  If queue is not empty, this function returns
  // immediately.
  GOTO_IF(error_unlock,
          wait_event_interruptible_lock_irq(q->wq, q->entry != NULL, q->lock));
  // Remove entry from queue.
  BUG_ON(q->entry == NULL);
  list_del(&q->entry->list_head);
  spin_unlock(&q->lock);

  *out = q->entry->raw;
  queue_entry_free(q->entry);
  return 0;

error_unlock:
  spin_unlock(&q->lock);
  return error;
}
