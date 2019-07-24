#include <sys/types.h>
#include <sys/xattr.h>
#include <attr/xattr.h>
#include <unistd.h>
#include <string.h>
#include "common.h"
#include "safe.h"

#define TEST_NAME "xattr"
#define ATTR_NAME "security.eltonfs-test"
#define VALUE_SIZE 256
#define SHORT_STRING "This is the short string to test xattr APIs."
#define SHORT_STRING_LEN strlen(SHORT_STRING)

#define DUMP_LIST(list, size) \
	do { \
		int i; \
		ssize_t pos; \
		for (i = 0, pos = 0; \
			pos < size; \
			i++, pos += strlen(&list[pos]) + 1) { \
			LOG_INFOF("list elm[%ld] = key:%s", i, &list[pos]); \
		} \
	}while(0)


void test_get_set_remove() {
	const int fd = SAFE_MKTEMP();
	char value[VALUE_SIZE];
	ssize_t size;

	LOG_INFO("Set an attribute (flags is not set)");
	CHECK_ERROR(fsetxattr(fd, ATTR_NAME, SHORT_STRING, SHORT_STRING_LEN, 0));

	LOG_INFO("Get an attribute");
	size = CHECK_ERROR(fgetxattr(fd, ATTR_NAME, value, VALUE_SIZE));
	EQUAL_INT(SHORT_STRING_LEN, size, "invalid size");

	LOG_INFO("Remove an attribute");
	CHECK_ERROR(fremovexattr(fd, ATTR_NAME));

	LOG_INFO("Lookup the not exist attribute");
	EQUAL_ERROR(ENOATTR, fgetxattr(fd, ATTR_NAME, value, VALUE_SIZE));
	close(fd);
}

void test_list() {
	const int fd = SAFE_MKTEMP();
	char list[VALUE_SIZE];
	ssize_t size;
	const size_t expected_list_size = strlen(ATTR_NAME "-" ATTR_NAME "2-");  // '-' means '\0'

	LOG_INFO("Get attribute names list");
	size = CHECK_ERROR(flistxattr(fd, list, VALUE_SIZE));
	EQUAL_INT(0, size, "invalid list size");

	LOG_INFO("Set an attribute");
	CHECK_ERROR(fsetxattr(fd, ATTR_NAME, SHORT_STRING, SHORT_STRING_LEN, 0));
	CHECK_ERROR(fsetxattr(fd, ATTR_NAME "2", SHORT_STRING, SHORT_STRING_LEN, 0));

	LOG_INFO("Try to get attribute names list with small buffer");
	EQUAL_ERROR(ERANGE, flistxattr(fd, list, 1));

	LOG_INFO("Get attribute names list");
	size = CHECK_ERROR(flistxattr(fd, list, VALUE_SIZE));
	LOG_INFOF("list size = %ld", size);
	DUMP_LIST(list, size);
	EQUAL_INT(expected_list_size, size, "invalid list size");
	close(fd);
}

void test_llist() {
	const char *from = TEMP_BASE "llist.from";
	const char *to = TEMP_BASE "llist.to";
	const char *attr_from = ATTR_NAME "-from";
	const char *attr_to = ATTR_NAME "-to";
	char list[VALUE_SIZE];
	ssize_t size;

	LOG_INFO("Create symlink");
	SAFE_MKFILE(from);
	CHECK_ERROR(symlink(from, to));

	LOG_INFO("Set attributes to link destination");
	CHECK_ERROR(setxattr(from, attr_from, SHORT_STRING, SHORT_STRING_LEN, 0));

	LOG_INFO("Set attributes to symlink itself");
	CHECK_ERROR(lsetxattr(to, attr_to, SHORT_STRING, SHORT_STRING_LEN, 0));

	LOG_INFO("Verify the llistxattr(2) response");
	size = CHECK_ERROR(llistxattr(to, list, VALUE_SIZE));
	DUMP_LIST(list, size);
	EQUAL_INT(strlen(attr_to)+1, size, "invalid list size");
}

int main(int argc, char **argv) {
	test_get_set_remove();
	test_list();
	test_llist();
	return 0;
}
