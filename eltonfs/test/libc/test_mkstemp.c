#include "common.h"
#include <stdio.h>
#include <stdlib.h>

int main(int argc, char **argv) {
  char template[] = "/mnt/mktemp-testXXXXXX";
  int fd = CHECK_ERROR(mkstemp(template));
  LOG_INFO(template);
  CHECK_ERROR(close(fd));
  return 0;
}
