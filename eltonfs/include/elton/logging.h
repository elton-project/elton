// Logging Macros
// ==============
//
// Constant Variables for Customize Logging Message:
//  * ELTON_LOG_PREFIX
//    A string literal of format string.
//  * ELTON_LOG_PREFIX_ARGS
//    Comma separated list of format args.  The args list MUST start with
//    a comma.
//
//  Usage:
//    #undef ELTON_LOG_PREFIX
//    #undef ELTON_LOG_PREFIX_ARGS
//    #define ELTON_LOG_PREFIX "session[id=%d]: "
//    #define ELTON_LOG_PREFIX_ARGS ,session->id
//
//
// Macros:
//  * DEBUG(format, ...)
//  * INFO(format, ...)
//  * ERR(format, ...)
//    Print log message using printk().

#ifndef _ELTON_LOGGING_H
#define _ELTON_LOGGING_H

#include <linux/printk.h>

#define ELTON_LOG_PREFIX
#define ELTON_LOG_PREFIX_ARGS

#define _PRINTLNK(level, fmt, ...)                                             \
  (printk(level MODULE_NAME ": " ELTON_LOG_PREFIX fmt                          \
                            "\n" ELTON_LOG_PREFIX_ARGS,                        \
          ##__VA_ARGS__))
#define DEBUG(fmt, ...) _PRINTLNK(KERN_DEBUG, fmt, ##__VA_ARGS__)
#define INFO(fmt, ...) _PRINTLNK(KERN_INFO, fmt, ##__VA_ARGS__)
#define ERR(fmt, ...) _PRINTLNK(KERN_ERR, fmt, ##__VA_ARGS__)

#endif // _ELTON_LOGGING_H
