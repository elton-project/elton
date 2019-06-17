#include <stdio.h>
#include <stdlib.h>
#include "common.h"

int main(int argc, char **argv) {
	char *name;
	name = mkdtemp("/mnt/mkdtemp-testXXXXXX");
	if(name == NULL) {
		PRINT_ERROR();
		exit(1);
	}
	LOG_INFO(name);
	return 0;
}
