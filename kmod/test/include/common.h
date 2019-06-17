#include <stdio.h>  // For perror()
#include <stdlib.h>  // For exit()
#include <unistd.h>  // For syscall()
#include <sys/syscall.h>  // For SYS_xxx constant values
#include <errno.h>
#include <fcntl.h>  // For O_xxx constant values

#define PRINT_ERROR() \
	do{ \
		perror("ERROR"); \
		fprintf(stderr, \
				"  Occurred on %s (%s:%d)\n", \
				__func__, __FILE__, __LINE__ \
				); \
	}while(0)

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

#define LOG_PRINTLN(level, msg) \
	fprintf(stderr, "%s: %s  (%s:%d %s)\n", level, msg, __FILE__, __LINE__, __func__);

#define LOG_INFO(msg) LOG_PRINTLN("INFO", msg)

