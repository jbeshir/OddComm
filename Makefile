include $(GOROOT)/src/Make.inc

PKGROOT = $(CURDIR)/pkg
PKGSEARCH = $(PKGROOT)/$(GOOS)_$(GOARCH)
PKGDIR = $(PKGSEARCH)/oddcomm

SUBSYSTEMS = $(patsubst %, $(PKGDIR)/src/%.a, $(shell cat subsystems.conf))
MODULES = $(patsubst %, $(PKGDIR)/modules/%.a, $(shell cat modules.conf))

GOCMD = $(GC) -I $(PKGSEARCH)
LDCMD = $(LD) -L $(PKGSEARCH)
GOPACK = $(GOBIN)/gopack

.PHONY:	all install clean


all:	oddcomm

clean:
	rm -rf $(PKGROOT)

oddcomm: $(SUBSYSTEMS) $(MODULES) $(PKGDIR)/src/core.a src/main/*.go
	$(GOCMD) -o oddcomm.$(O) $(wildcard src/main/*.go)
	rm -f oddcomm
	$(LDCMD) -o oddcomm oddcomm.$(O)
	rm -f oddcomm.$(O)

$(PKGDIR)/modules/%.a: $(SUBSYSTEMS) $(PKGDIR)/src/core.a modules/%/*.go
	mkdir -p $(PKGDIR)/modules/$*
	rmdir $(PKGDIR)/modules/$*
	$(GOCMD) -o $(PKGDIR)/modules/$*.$(O) $(wildcard modules/$*/*.go)
	rm -f $(PKGDIR)/modules/$*.a
	$(GOPACK) grc $(PKGDIR)/modules/$*.a $(PKGDIR)/modules/$*.$(O)
	rm -f $(PKGDIR)/modules/$*.$(O)

$(PKGDIR)/src/client.a: $(PKGDIR)/src/core.a $(PKGDIR)/lib/irc.a $(PKGDIR)/lib/perm.a src/client/*.go
	mkdir -p $(PKGDIR)/src
	$(GOCMD) -o $(PKGDIR)/src/client.$(O) $(wildcard src/client/*.go)
	rm -f $(PKGDIR)/src/client.a
	$(GOPACK) grc $(PKGDIR)/src/client.a $(PKGDIR)/src/client.$(O)
	rm -f $(PKGDIR)/src/client.$(O)

$(PKGDIR)/lib/%.a: $(PKGDIR)/src/core.a lib/%/*.go
	mkdir -p $(PKGDIR)/lib
	$(GOCMD) -o $(PKGDIR)/lib/$*.$(O) $(wildcard lib/$*/*.go)
	rm -f $(PKGDIR)/lib/$*.a
	$(GOPACK) grc $(PKGDIR)/lib/$*.a $(PKGDIR)/lib/$*.$(O)
	rm -f $(PKGDIR)/lib/$*.$(O)

$(PKGDIR)/src/core.a: src/core/*.go $(PKGDIR)/lib/trie.a
	mkdir -p $(PKGDIR)/src
	$(GOCMD) -o $(PKGDIR)/src/core.$(O) $(wildcard src/core/*.go)
	rm -f $(PKGDIR)/src/core.a
	$(GOPACK) grc $(PKGDIR)/src/core.a $(PKGDIR)/src/core.$(O)
	rm -f $(PKGDIR)/src/core.$(O)

$(PKGDIR)/lib/trie.a: lib/trie/*.go
	mkdir -p $(PKGDIR)/lib
	$(GOCMD) -o $(PKGDIR)/lib/trie.$(O) $(wildcard lib/trie/*.go)
	rm -f $(PKGDIR)/lib/trie.a
	$(GOPACK) grc $(PKGDIR)/lib/trie.a $(PKGDIR)/lib/trie.$(O)
	rm -f $(PKGDIR)/lib/trie.$(O)

