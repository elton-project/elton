
// Create a temporary file and return fd.
#define SAFE_MKTEMP() \
	({ \
		char name[] = "/mnt/test_" TEST_NAME "XXXXXX"; \
		int fd = CHECK_ERROR(mkstemp(name)); \
		fd; \
	})
