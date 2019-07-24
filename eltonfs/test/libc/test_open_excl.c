#include <stdio.h>
#include <stdlib.h>
#include "common.h"

int main(int argc, char **argv) {
	int fd;

	LOG_INFO("Creating new file...");
	fd = CHECK_ERROR(open("/mnt/open_excl-new-file", O_EXCL|O_CREAT|O_RDWR));
	CHECK_ERROR(close(fd));
	LOG_INFO("OK");

	LOG_INFO("Try to create with exists file (This test case should fail)");
	fd = open("/mnt/open_excl-new-file", O_EXCL|O_CREAT|O_RDWR);
	ASSERT(fd == -1, "should fail, but succeed")
	LOG_INFO("OK");
	return 0;
}
