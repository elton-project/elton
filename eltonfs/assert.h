#ifndef _ELTON_ASSERT_H
#define _ELTON_ASSERT_H

#include <linux/printk.h>
#include <elton.h>

// ASSERT_*()マクロが失敗したとき、trueに設定される。
// デフォルト値はfalse。
extern volatile bool __assertion_failed;


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
// ポインタがNULLの場合、WARNINGを表示する。
#define ASSERT_NOT_NULL(expr) ({ \
	typeof(expr) error = expr; \
	bool fail = !error; \
	if(fail) { \
		ERR("ASSERT: %s is NULL (%s %s:%d)", #expr, __func__, __FILE__, __LINE__); \
		SET_ASSERTION_FAILED(); \
	} \
	fail; \
})
#define ASSERT_NO_ERROR(expr) ({ \
	typeof(expr) error = expr;\
	bool fail = error; \
	if(fail) { \
		ERR("ASSERT: %s returns an error %d (%s %s:%d)", #expr, error, __func__, __FILE__, __LINE__); \
	} \
	fail; \
})
#define ASSERT_EQUAL_ERROR(expected, expr) ({ \
	typeof(expr) actual = expr; \
	bool fail = actual != expected; \
	if(fail) { \
		ERR("ASSERT: %s returns unexpected error (%s %s:%d)\n" \
			"  expected=%d\n" \
			"  actual=%d", \
			#expr, __func__, __FILE__, __LINE__, \
			expected, actual); \
	} \
	fail; \
})
// Assertionが失敗したことにする。
#define SET_ASSERTION_FAILED() do{ __assertion_failed = true; }while(0)
// Assertionが失敗したときにtrueを返す。
#define IS_ASSERTION_FAILED() (!!__assertion_failed)


#endif // _ELTON_ASSERT_H
