// The base path for temporary file.
#define TEMP_BASE "/mnt/test_" TEST_NAME "."
// The template for temporary file path.
#define TEMP_PATTERN TEMP_BASE "XXXXXX"

// Create a temporary file and return fd.
#define SAFE_MKTEMP() \
	({ \
		char name[] = TEMP_PATTERN; \
		int fd = CHECK_ERROR(mkstemp(name)); \
		fd; \
	})
