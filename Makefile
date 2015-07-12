TARGETDIR = bin

all: client eltonfs

client:
	@$(MAKE) -C cmd

eltonfs:
	@$(MAKE) -C eltonfs

clean:
	$(RM) -r $(TARGETDIR)

.PHONY: all client eltonfs clean
