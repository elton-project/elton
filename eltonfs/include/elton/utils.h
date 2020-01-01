#include <elton/assert.h>
#include <elton/elton.h>
#include <linux/kthread.h>
#include <linux/types.h>

// Run a function in background asynchronous/synchronous.
//
// Functions:
//   eltonfs_job_run(...);
//   eltonfs_job_run_sync(...);
//   eltonfs_job_wait(...);
//
// Usage:
//   struct eltonfs_job job;
//   void *output;
//   RETURN_IF(eltonfs_job_run(&job, fn, input, "example"));
//   output = eltonfs_job_wait(&job));
//   if(IS_ERR(output))
//       RETURN_IF(PTR_ERR(output));
struct eltonfs_job {
  struct completion done;
  void *(*fn)(void *input);
  void *input;
  void *output;
  struct task_struct *task;
};

static __maybe_unused int _eltonfs_job_worker(void *_job) {
  struct eltonfs_job *job = (struct eltonfs_job *)_job;
  job->output = job->fn(job->input);
  complete(&job->done);
  return 0;
}
static __maybe_unused void _eltonfs_job_init(struct eltonfs_job *job,
                                             void *(*fn)(void *input),
                                             void *input) {
  init_completion(&job->done);
  job->fn = fn;
  job->input = input;
  job->output = NULL;
}
static __maybe_unused int eltonfs_job_run(struct eltonfs_job *job,
                                          void *(*fn)(void *input), void *input,
                                          const char *name) {
  struct task_struct *task;
  _eltonfs_job_init(job, fn, input);
  task = kthread_run(_eltonfs_job_worker, job, "eltonfs-%s", name);
  if (IS_ERR(task))
    return PTR_ERR(task);
  job->task = task;
  return 0;
}
static __maybe_unused void *eltonfs_job_wait(struct eltonfs_job *job) {
  int error = wait_for_completion_interruptible(&job->done);
  if (error)
    return ERR_PTR(error);
  kthread_stop(job->task);
  return job->output;
}
static __maybe_unused void *
eltonfs_job_run_sync(void *(*fn)(void *input), void *input, const char *name) {
  int error;
  struct eltonfs_job job;
  error = eltonfs_job_run(&job, fn, input, name);
  if (error)
    return ERR_PTR(error);
  return eltonfs_job_wait(&job);
}

// Duplicate NULL terminated string.
static inline __maybe_unused int dup_string(char **to, const char *from) {
  int error = 0;
  size_t len = strlen(from);
  char *buff = kmalloc(len + 1, GFP_NOFS);
  if (!buff)
    RETURN_IF(-ENOMEM);
  strcpy(buff, from);
  *to = buff;
  return 0;
}

// Initialize "to" list and copies list contents.
// All entries are shallow copied.
static inline __maybe_unused int
dup_dir_entries(struct eltonfs_dir_entry *to,
                const struct eltonfs_dir_entry *from) {
  const struct eltonfs_dir_entry *entry;
  struct eltonfs_dir_entry *copy;

  INIT_LIST_HEAD(&to->_list_head);
  list_for_each_entry(entry, &from->_list_head, _list_head) {
    copy = kmalloc(sizeof(*copy), GFP_NOFS);
    if (!copy)
      return -ENOMEM;
    memcpy(copy, entry, sizeof(*copy));
    list_add_tail(&copy->_list_head, &to->_list_head);
  }
  return 0;
}
