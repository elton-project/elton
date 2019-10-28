#ifndef _ELTON_COMPILER_ATTRIBUTES_H
#define _ELTON_COMPILER_ATTRIBUTES_H

#ifdef __GNUC__
#define __unused __attribute__((unused))
#else // __GNUC__
#define __unused
#endif

#endif // _ELTON_COMPILER_ATTRIBUTES_H
