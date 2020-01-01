#include <elton/assert.h>
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
