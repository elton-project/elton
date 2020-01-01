#include <elton/assert.h>
#include <elton/elton.h>
#include <linux/types.h>

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
