#include <sys/mman.h>
#include <string.h>
#include "common.h"

#define FILE_NAME "/mnt/mmap.bin"
#define DATA_LENGTH 10000
#define DATA_TYPE int
#define FILE_SIZE sizeof(DATA_TYPE)*DATA_LENGTH

int fill_zero(int fd, int bytes) {
	int i;
	char buf[1024];
	memset(buf, 0, 1024);
	for(i=0; i<bytes; i+=1024)
		write(fd, buf, 1024);
}

int main(int argc, char **argv) {
	int fd;
	int page_size;
	int map_size;
	void *map;
	DATA_TYPE i;
	DATA_TYPE *p;

	LOG_INFO("creating a file");
	unlink(FILE_NAME);
	fd = CHECK_ERROR(open(FILE_NAME, O_CREAT|O_EXCL|O_RDWR));
	page_size = getpagesize();
	map_size = (FILE_SIZE / page_size + 1) * page_size;
	fill_zero(fd, map_size);
	lseek(fd, 0, SEEK_SET);


	LOG_INFO("writing bytes to mmaped area");
	p = map = mmap(NULL, map_size, PROT_WRITE, MAP_SHARED, fd, 0);
	ASSERT(map != MAP_FAILED, "mmap failed");
	for(i=0; i<DATA_LENGTH; i++)
		p[i] = i;
	CHECK_ERROR(munmap(map, map_size));


	LOG_INFO("reading bytes from mmaped area");
	p = map = mmap(NULL, map_size, PROT_READ, MAP_SHARED, fd, 0);
	ASSERT(map != MAP_FAILED, "mmap failed");
	for(i=0; i<DATA_LENGTH; i++)
		ASSERT(p[i] == i, "incorrect data");
	CHECK_ERROR(munmap(map, map_size));


	LOG_INFO("removing a file");
	CHECK_ERROR(close(fd));
	CHECK_ERROR(unlink(FILE_NAME));
	return 0;
}
