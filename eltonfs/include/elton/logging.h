// Logging Macros
// ==============
//
// Macros for Customize Logging Message:
//  * ELTON_LOG_PREFIX
//    A string literal of format string.
//  * ELTON_LOG_PREFIX_ARGS
//    Comma separated list of format args.  The args list MUST start with
//    a comma.
//
//  Usage:
//    /* Define macros before include the <elton/logging.h> */
//    #define ELTON_LOG_PREFIX "session[id=%d]: "
//    #define ELTON_LOG_PREFIX_ARGS ,session->id
//
//    #include <header/files.h>
//    void foo() {
//      INFO("called foo()");
//    }
//
//
// Macros for Printing Log Messages:
//  * DEBUG(format, ...)
//  * INFO(format, ...)
//  * ERR(format, ...)
//    Print log message using printk().

#ifndef _ELTON_LOGGING_H
#define _ELTON_LOGGING_H

#include <linux/printk.h>

#ifndef ELTON_LOG_PREFIX
#define ELTON_LOG_PREFIX
#endif
#ifndef ELTON_LOG_PREFIX_ARGS
#define ELTON_LOG_PREFIX_ARGS
#endif

#define _PRINTLNK(level, fmt, ...)                                             \
  (printk(level MODULE_NAME ": " ELTON_LOG_PREFIX fmt                          \
                            "\n" ELTON_LOG_PREFIX_ARGS,                        \
          ##__VA_ARGS__))
#define DEBUG(fmt, ...) _PRINTLNK(KERN_DEBUG, fmt, ##__VA_ARGS__)
#define INFO(fmt, ...) _PRINTLNK(KERN_INFO, fmt, ##__VA_ARGS__)
#define ERR(fmt, ...) _PRINTLNK(KERN_ERR, fmt, ##__VA_ARGS__)

#endif // _ELTON_LOGGING_H
