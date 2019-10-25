#ifndef _ELTON_ASSERT_H
#define _ELTON_ASSERT_H

#include <linux/printk.h>
#include <elton.h>


#define _PRINTLNK(level, fmt, ...) (printk(level MODULE_NAME ": " fmt "\n", ##__VA_ARGS__))
#define DEBUG(fmt, ...) _PRINTLNK(KERN_DEBUG, fmt, ##__VA_ARGS__)
#define INFO(fmt, ...) _PRINTLNK(KERN_INFO, fmt, ##__VA_ARGS__)
#define ERR(fmt, ...) _PRINTLNK(KERN_ERR, fmt, ##__VA_ARGS__)

// エラーならtrueを返す。
// また、エラー発生時にログを残す。
#define CHECK_ERROR(expr) ({ \
    typeof(expr) error = expr; \
	if(error) { \
		ERR("Occurred an error %d on %s (%s:%d)", error, __func__, __FILE__, __LINE__); \
	} \
	error; \
})
#define ASSERT_NOT_NULL(expr) ({ \
	typeof(expr) error = expr; \
	if(!expr) { \
		ERR(#expr " is NULL"); \
		BUG_ON(p); \
	} \
	error; \
})


#endif // _ELTON_ASSERT_H
