#include "common.h"
#include <stdio.h>
#include <stdlib.h>

int main(int argc, char **argv) {
  char template[] = "/mnt/mkdtemp-testXXXXXX";
  char *name = mkdtemp(template);
  if (name == NULL) {
    PRINT_ERROR();
    exit(1);
  }
  LOG_INFO(name);
  return 0;
}
