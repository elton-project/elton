TEST_CACES = $(patsubst %.c,%.run,$(wildcard test_*.c))

.PHONY: test_all
test_all: $(TEST_CACES)

.SUFFIXES: .out .run
.out.run:
	@echo '================ start $< ================'
	./$<
	@echo '================ end $< ================'

.SUFFIXES: .c .out
.c.out:
	gcc -ggdb -I../include -o $@ $<
