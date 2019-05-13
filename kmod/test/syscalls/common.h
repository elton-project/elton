#include <stdio.h>  // For perror()
#include <stdlib.h>  // For exit()
#include <unistd.h>  // For syscall()
#include <sys/syscall.h>  // For SYS_xxx constant values
#include <errno.h>
#include <fcntl.h>  // For O_xxx constant values

#define CHECK_ERROR(expr) \
	({ \
		int ret; \
		errno = 0; \
		ret = (expr); \
		if(ret == -1 && errno) { \
			perror("ERROR"); \
			fprintf(stderr, \
					"  Occurred on %s (%s:%d)\n" \
					"  Expr: %s\n", \
					__func__, __FILE__, __LINE__, \
					#expr \
					); \
			exit(1); \
		} \
		ret; \
	})

#define ASSERT(expr, msg) \
	if(! (expr)) { \
		fprintf(stderr, \
				"ERROR: %s\n" \
				"  Occurred on %s (%s:%d)\n" \
				"  Expr: %s\n", \
				msg, \
				__func__, __FILE__, __LINE__, \
				#expr \
				); \
		exit(1); \
	}

