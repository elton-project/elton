#ifndef _ELTON_ASSERT_H
#define _ELTON_ASSERT_H

#include <elton/elton.h>
#include <linux/bug.h>
#include <linux/printk.h>

#define lengthof(array) (sizeof(array) / sizeof((array)[0]))
// Break compile if array size is not match.
#define BUILD_ASSERT_EQUAL_ARRAY_SIZE(expected, array)                         \
  BUILD_BUG_ON((expected) != lengthof(array))

// ASSERT_*()マクロが失敗したとき、trueに設定される。
// デフォルト値はfalse。
extern volatile bool __assertion_failed;

#define _PRINTLNK(level, fmt, ...)                                             \
  (printk(level MODULE_NAME ": " fmt "\n", ##__VA_ARGS__))
#define DEBUG(fmt, ...) _PRINTLNK(KERN_DEBUG, fmt, ##__VA_ARGS__)
#define INFO(fmt, ...) _PRINTLNK(KERN_INFO, fmt, ##__VA_ARGS__)
#define ERR(fmt, ...) _PRINTLNK(KERN_ERR, fmt, ##__VA_ARGS__)

// エラーならtrueを返す。
// また、エラー発生時にログを残す。
#define CHECK_ERROR(expr)                                                      \
  ({                                                                           \
    typeof(expr) result = expr;                                                \
    bool is_error = result < 0;                                                \
    if (is_error) {                                                            \
      ERR("Occurred an error %d on %s (%s:%d)", (int)result, __func__,         \
          __FILE__, __LINE__);                                                 \
    }                                                                          \
    is_error;                                                                  \
  })
// エラーなら指定したラベルにジャンプする。
#define GOTO_IF(label, expr)                                                   \
  do {                                                                         \
    error = (expr);                                                            \
    if (CHECK_ERROR(error))                                                    \
      goto label;                                                              \
  } while (0)
// エラーならreturnする。関数に戻り値がない場合、RETURN_VOID_IF()を使う。
#define RETURN_IF(expr)                                                        \
  do {                                                                         \
    error = (expr);                                                            \
    if (CHECK_ERROR(error))                                                    \
      return error;                                                            \
  } while (0)
// エラーならreturnする。戻り値が無い関数用。
#define RETURN_VOID_IF(expr)                                                   \
  do {                                                                         \
    error = (expr);                                                            \
    if (CHECK_ERROR(error))                                                    \
      return;                                                                  \
  } while (0)
// ポインタがNULLの場合、WARNINGを表示してtrueを返す。
// それ以外の場合は、falseを返す。
#define ASSERT_NOT_NULL(expr)                                                  \
  ({                                                                           \
    typeof(expr) error = expr;                                                 \
    bool fail = !error;                                                        \
    if (fail) {                                                                \
      ERR("ASSERT: %s is NULL (%s %s:%d)", #expr, __func__, __FILE__,          \
          __LINE__);                                                           \
      SET_ASSERTION_FAILED();                                                  \
    }                                                                          \
    fail;                                                                      \
  })
// 条件式がエラーを返した場合、WARNINGを表示してtrueを返す。
// それ以外の場合は、falseを返す。
#define ASSERT_NO_ERROR(expr)                                                  \
  ({                                                                           \
    typeof(expr) error = expr;                                                 \
    bool fail = error;                                                         \
    if (fail) {                                                                \
      ERR("ASSERT: %s returns an error %d (%s %s:%d)", #expr, error, __func__, \
          __FILE__, __LINE__);                                                 \
      SET_ASSERTION_FAILED();                                                  \
    }                                                                          \
    fail;                                                                      \
  })
// 条件式が指定したエラーコードと異なる値を返した場合、WARNINGを表示してtrueを返す。
// それ以外の場合は、falseを返す。
#define ASSERT_EQUAL_ERROR(expected, expr)                                     \
  ({                                                                           \
    typeof(expr) actual = expr;                                                \
    bool fail = actual != expected;                                            \
    if (fail) {                                                                \
      ERR("ASSERT: %s returns unexpected error (%s %s:%d): "                   \
          "expected=%d actual=%d",                                             \
          #expr, __func__, __FILE__, __LINE__, expected, actual);              \
      SET_ASSERTION_FAILED();                                                  \
    }                                                                          \
    fail;                                                                      \
  })
// 条件式が想定外の数値を返した場合、WARNINGを表示してtrueを返す。
// それ以外の場合は、falseを返す。
#define ASSERT_EQUAL_INT(expected, expr)                                       \
  ({                                                                           \
    int actual = expr;                                                         \
    bool fail = actual != expected;                                            \
    if (fail) {                                                                \
      ERR("ASSERT: %s return expeccted value (%s %s:%d): "                     \
          "expected=%d actual=%d",                                             \
          #expr, __func__, __FILE__, __LINE__, expected, actual);              \
      SET_ASSERTION_FAILED();                                                  \
    }                                                                          \
    fail;                                                                      \
  })
// 条件式が想定外の数値を返した場合、WARNINGを表示してtrueを返す。
// それ以外の場合は、falseを返す。
#define ASSERT_EQUAL_LL(expected, expr)                                        \
  ({                                                                           \
    long long actual = expr;                                                   \
    bool fail = actual != expected;                                            \
    if (fail) {                                                                \
      ERR("ASSERT: %s return expeccted value (%s %s:%d): "                     \
          "expected=%lld actual=%lld",                                         \
          #expr, __func__, __FILE__, __LINE__, expected, actual);              \
      SET_ASSERTION_FAILED();                                                  \
    }                                                                          \
    fail;                                                                      \
  })
// 条件式が想定外のサイズを返した場合、WARNINGを表示してtrueを返す。
// それ以外の場合は、falseを返す。
#define ASSERT_EQUAL_SIZE_T(expected, expr)                                    \
  ({                                                                           \
    size_t actual = expr;                                                      \
    bool fail = actual != expected;                                            \
    if (fail) {                                                                \
      ERR("ASSERT: %s return unexpected value (%s %s:%d): "                    \
          "expected=%zu actual=%zu",                                           \
          #expr, __func__, __FILE__, __LINE__, expected, actual);              \
      SET_ASSERTION_FAILED();                                                  \
    }                                                                          \
    fail;                                                                      \
  })
// 2つのchar*型の配列の内容が一致しない場合、WARNINGを表示してtrueを返す。
// それ以外の場合は、falseを返す。
// 表示内容は、
#define ASSERT_EQUAL_BYTES(expected, actual, size)                             \
  ({                                                                           \
    int i, result = memcmp(expected, actual, size);                            \
    if (result) {                                                              \
      ERR("ASSERT: two values does not match (%s %s:%d)", __func__, __FILE__,  \
          __LINE__);                                                           \
      ERR("ASSERT:    expected  actual");                                      \
      for (i = 0; i < size; i++) {                                             \
        char e = (expected)[i], a = (actual)[i];                               \
        if (e == a)                                                            \
          ERR("ASSERT:  [%d]  %d   %d", i, e, a);                              \
        else                                                                   \
          ERR("ASSERT:  [%d]  %d   %d   <=== !!", i, e, a);                    \
      }                                                                        \
      SET_ASSERTION_FAILED();                                                  \
    }                                                                          \
    result;                                                                    \
  })
// Assertionが失敗したことにする。
#define SET_ASSERTION_FAILED()                                                 \
  do {                                                                         \
    __assertion_failed = true;                                                 \
  } while (0)
// Assertionが失敗したときにtrueを返す。
#define IS_ASSERTION_FAILED() (!!__assertion_failed)

#endif // _ELTON_ASSERT_H
