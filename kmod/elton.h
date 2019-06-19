#ifndef _ELTON_ELTON_H
#define _ELTON_ELTON_H


#define MODULE_NAME "elton"
#define FS_NAME "elton"
#define ELTONFS_SUPER_MAGIC 0x51891f5
#define ELTONFS_NAME_LEN 255


struct eltonfs_info {
#ifdef ELTONFS_STATISTIC
	unsigned long mmap_size;
	rwlock_t mmap_size_lock;
#endif
};


#endif // _ELTON_ELTON_H
