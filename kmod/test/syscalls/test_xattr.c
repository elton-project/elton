#include <sys/types.h>
#include <sys/xattr.h>
#include <attr/xattr.h>
#include <string.h>
#include "common.h"
#include "safe.h"

#define TEST_NAME "xattr"
#define ATTR_NAME "user.attr"
#define VALUE_SIZE 256
#define SHORT_STRING "This is the short string to test xattr APIs."
#define SHORT_STRING_LEN strlen(SHORT_STRING)

void test_get_set_remove() {
	int fd = SAFE_MKTEMP();
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
	int fd = SAFE_MKTEMP();
	char list[VALUE_SIZE];
	ssize_t size;

	LOG_INFO("Get attribute names list");
	size = CHECK_ERROR(flistxattr(fd, list, VALUE_SIZE));
	EQUAL_INT(0, size, "invalid list size");

	LOG_INFO("Set an attribute");
	CHECK_ERROR(fsetxattr(fd, ATTR_NAME, SHORT_STRING, SHORT_STRING_LEN, 0));
	CHECK_ERROR(fsetxattr(fd, ATTR_NAME "2", SHORT_STRING, SHORT_STRING_LEN, 0));

	LOG_INFO("Get attribute names list");
	size = CHECK_ERROR(flistxattr(fd, list, VALUE_SIZE));
	EQUAL_INT(2, size, "invalid list size");
	close(fd);
}

int main(int argc, char **argv) {
	test_get_set_remove();
	test_list();
	return 0;
}
