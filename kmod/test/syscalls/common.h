#include <stdio.h>  // For perror()
#include <stdlib.h>  // For exit()
#include <unistd.h>  // For syscall()
#include <sys/syscall.h>  // For SYS_xxx constant values
#include <errno.h>
#include <fcntl.h>  // For O_xxx constant values

#define CHECK_ERROR(expr) \
	do{ \
		int ret; \
		errno = 0; \
		ret = (expr); \
		if(ret && errno) { \
			perror("ERROR"); \
			exit(1); \
		} \
	}while(0)

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

