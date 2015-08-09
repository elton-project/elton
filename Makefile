TARGETDIR = bin

all: client eltonfs

client:
	@$(MAKE) -C cmd

eltonfs:
	@$(MAKE) -C eltonfs

install:
	@$(MAKE) -C cmd install
	@$(MAKE) -C eltonfs install

clean:
	$(RM) -r $(TARGETDIR)

.PHONY: all client install eltonfs clean
