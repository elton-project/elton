#include <stdio.h>
#include <stdlib.h>
#include "common.h"

int main(int argc, char **argv) {
	char *name;
	name = CHECK_ERROR(mkdtemp("/mnt/mkdtemp-testXXXXXX"));
	LOG_INFO(name);
	return 0;
}
