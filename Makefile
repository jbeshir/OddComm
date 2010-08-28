include $(GOROOT)/src/Make.inc

DESTDIR = .
PKGROOT = $(CURDIR)/pkg
PKGSEARCH = $(PKGROOT)/$(GOOS)_$(GOARCH)
PKGDIR = $(PKGSEARCH)/oddcomm

SUBSYSTEMS = $(patsubst %, $(PKGDIR)/src/%.a, $(shell cat subsystems.conf))
MODULES = $(patsubst %, $(PKGDIR)/modules/%.a, $(shell cat modules.conf))

GOCMD = $(GC) -I $(PKGSEARCH)
LDCMD = $(LD) -L $(PKGSEARCH)

.PHONY:	all install clean


all:	install

install: src/main/oddcomm
	install src/main/oddcomm $(DESTDIR)

clean:
	rm -rf $(PKGROOT)
	rm -f src/main/oddcomm

src/main/oddcomm: $(SUBSYSTEMS) $(MODULES) $(PKGDIR)/src/core.a src/main/*.go
	$(GOCMD) -o src/main/_go_.$(O) $(wildcard src/main/*.go)
	rm -f src/main/oddcomm
	$(LDCMD) -o src/main/oddcomm src/main/_go_.$(O)
	rm -f src/main/_go_.$(O)

$(PKGDIR)/modules/%.a: $(SUBSYSTEMS) $(PKGDIR)/src/core.a modules/%/*.go
	$(GOCMD) -o modules/$*/_go_.$(O) $(wildcard modules/$*/*.go)
	mkdir -p $(PKGDIR)/modules/$*
	rmdir $(PKGDIR)/modules/$*
	gopack grc $(PKGDIR)/modules/$*.a modules/$*/_go_.$(O)
	rm -f modules/$*/_go_.$(O)

$(PKGDIR)/src/client.a: $(PKGDIR)/src/core.a $(PKGDIR)/lib/irc.a $(PKGDIR)/lib/perm.a src/client/*.go
	$(GOCMD) -o src/client/_go_.$(O) $(wildcard src/client/*.go)
	mkdir -p $(PKGDIR)/src
	gopack grc $(PKGDIR)/src/client.a src/client/_go_.$(O)
	rm -f src/client/_go_.$(O)

$(PKGDIR)/lib/%.a: $(PKGDIR)/src/core.a lib/%/*go
	$(GOCMD) -o lib/$*/_go_.$(O) $(wildcard lib/$*/*.go)
	mkdir -p $(PKGDIR)/lib
	gopack grc $(PKGDIR)/lib/$*.a lib/$*/_go_.$(O)
	rm -f lib/$*/_go_.$(O)

$(PKGDIR)/src/core.a: src/core/*.go
	$(GOCMD) -o src/core/_go_.$(O) $(wildcard src/core/*.go)
	mkdir -p $(PKGDIR)/src
	rm -f $(PKGDIR)/src/core.a
	gopack grc $(PKGDIR)/src/core.a src/core/_go_.$(O)
	rm -f src/core/_go_.$(O)
