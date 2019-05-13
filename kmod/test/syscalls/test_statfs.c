#include <sys/vfs.h>
#include "common.h"

int main(int argc, char **argv) {
	struct statfs buf;
	statfs("/mnt", &buf);
	CHECK_ERROR();
	printf(
		"f_type = 0x%lx\n"
		"f_bsize = %ld\n"
		"f_blocks = %ld\n"
		"f_bfree = %ld\n"
		"f_bavail = %ld\n"
		"f_files = %ld\n"
		"f_ffree = %ld\n",
		buf.f_type,
		buf.f_bsize,
		buf.f_blocks,
		buf.f_bfree,
		buf.f_bavail,
		buf.f_files,
		buf.f_ffree
	);

	{
		char *msg = "some statfs fields are 0";
		ASSERT(buf.f_type, msg);
		ASSERT(buf.f_bsize, msg);
		ASSERT(buf.f_blocks, msg);
		ASSERT(buf.f_bfree, msg);
		ASSERT(buf.f_bavail, msg);
		ASSERT(buf.f_files, msg);
		ASSERT(buf.f_ffree, msg);
	}
	return 0;
}
