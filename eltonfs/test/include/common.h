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

#define EQUAL_ERROR(expect, expr) \
	({ \
		int ret; \
		int ok = 0; \
		errno = 0; \
		ret = (expr); \
		if(ret == -1 && errno==expect) ok = 1; \
		else if(ret == -1 && errno) { \
			perror("ERROR"); \
			fprintf(stderr, \
					"  Occurred on %s (%s:%d)\n" \
					"  Expr: %s\n", \
					__func__, __FILE__, __LINE__, \
					#expr \
					); \
			exit(1); \
	    }else{ \
			fprintf(stderr, \
					"Received unexpected value: %d\n" \
					"  Occurred on %s (%s:%d)\n" \
					"  Expr: %s\n", \
					ret, \
					__func__, __FILE__, __LINE__, \
					#expr \
					); \
			exit(1); \
		} \
		ok; \
	})

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
				"ASSERT: %s\n" \
				"  Occurred on %s (%s:%d)\n" \
				"  Expr: %s\n", \
				msg, \
				__func__, __FILE__, __LINE__, \
				#expr \
				); \
		exit(1); \
	}

#define EQUAL_INT(expect, expr, msg) ({ \
		int ret; \
		ret = (expr); \
		if((expect) != ret) { \
			fprintf(stderr, \
					"ASSERT: %s\n" \
					"  Occurred on %s (%s:%d)\n" \
					"  Expr: %s\n" \
					"  Result: %d\n" \
					"  Expected: %ld\n", \
					msg, \
					__func__, __FILE__, __LINE__, \
					#expr, \
					ret, \
					(long int)(expect) \
					); \
			exit(1); \
		} \
	})

#define LOG_PRINTLN(level, msg) \
	fprintf(stderr, "%s: %s  (%s:%d %s)\n", level, msg, __FILE__, (int)(__LINE__), __func__);
#define LOG_PRINTF(level, msg, ...) \
	fprintf(stderr, "%s: " msg "  (%s:%d %s)\n", level, __VA_ARGS__, __FILE__, (int)(__LINE__), __func__);

#define LOG_INFO(msg) LOG_PRINTLN("INFO", msg)
#define LOG_INFOF(msg, ...) LOG_PRINTF("INFO", msg, __VA_ARGS__)

