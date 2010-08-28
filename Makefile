DESTDIR := .
PKGROOT := $(CURDIR)/pkg
PKGSEARCH := $(PKGROOT)/$(GOOS)_$(GOARCH)
PKGDIR := $(PKGSEARCH)/oddcomm

SUBSYSTEMS := $(patsubst %, $(PKGDIR)/src/%.a, $(shell cat subsystems.conf))
MODULES := $(patsubst %, $(PKGDIR)/modules/%.a, $(shell cat modules.conf))

ARCH:= 6
GOCMD := $(ARCH)g -I $(PKGSEARCH)
LDCMD := $(ARCH)l -L $(PKGSEARCH)

.PHONY:	all install clean


all:	install

install: src/main/oddcomm
	install src/main/oddcomm $(DESTDIR)

clean:
	rm -rf $(PKGROOT)

src/main/oddcomm: $(SUBSYSTEMS) $(MODULES) $(PKGDIR)/src/core.a src/main/*.go
	$(GOCMD) -o src/main/_go_.$(ARCH) $(wildcard src/main/*.go)
	rm -f src/main/oddcomm
	$(LDCMD) -o src/main/oddcomm src/main/_go_.6
	rm -f src/main/_go_.$(ARCH)

$(PKGDIR)/modules/%.a: $(SUBSYSTEMS) $(PKGDIR)/src/core.a modules/%/*.go
	$(GOCMD) -o modules/$*/_go_.$(ARCH) $(wildcard modules/$*/*.go)
	mkdir -p $(PKGDIR)/modules/$*
	rmdir $(PKGDIR)/modules/$*
	gopack grc $(PKGDIR)/modules/$*.a modules/$*/_go_.6
	rm -f modules/$*/_go_.$(ARCH)

$(PKGDIR)/src/client.a: $(PKGDIR)/src/core.a $(PKGDIR)/lib/irc.a $(PKGDIR)/lib/perm.a src/client/*.go
	$(GOCMD) -o src/client/_go_.$(ARCH) $(wildcard src/client/*.go)
	mkdir -p $(PKGDIR)/src
	gopack grc $(PKGDIR)/src/client.a src/client/_go_.6
	rm -f src/client/_go_.$(ARCH)

$(PKGDIR)/lib/%.a: $(PKGDIR)/src/core.a lib/%/*go
	$(GOCMD) -o lib/$*/_go_.$(ARCH) $(wildcard lib/$*/*.go)
	mkdir -p $(PKGDIR)/lib
	gopack grc $(PKGDIR)/lib/$*.a lib/$*/_go_.6
	rm -f lib/$*/_go_.$(ARCH)

$(PKGDIR)/src/core.a: src/core/*.go
	$(GOCMD) -o src/core/_go_.$(ARCH) $(wildcard src/core/*.go)
	mkdir -p $(PKGDIR)/src
	rm -f $(PKGDIR)/src/core.a
	gopack grc $(PKGDIR)/src/core.a src/core/_go_.6
	rm -f src/core/_go_.$(ARCH)
