ifndef $(GOROOT)
	GOROOT=$(HOME)/go
endif

include $(GOROOT)/src/Make.inc

PKGROOT = $(CURDIR)/pkg
PKGSEARCH = $(PKGROOT)/$(GOOS)_$(GOARCH)
PKGDIR = $(PKGSEARCH)/oddcomm

CORE = $(PKGDIR)/src/core.a $(PKGDIR)/lib/trie.a $(PKGDIR)/lib/cas.a
SUBSYSTEMS = $(patsubst %, $(PKGDIR)/src/%.a, $(shell cat subsystems.conf))
MODULES = $(patsubst %, $(PKGDIR)/modules/%.a, $(shell cat modules.conf))

GOASM = $(GOBIN)/$(O)a
GOCMD = $(GOBIN)/$(GC) -I $(PKGSEARCH)
LDCMD = $(GOBIN)/$(LD) -L $(PKGSEARCH)
GOFMT = $(GOBIN)/gofmt
GOPACK = $(GOBIN)/gopack

.PHONY:	all install clean


all:	oddcomm

clean:
	rm -rf $(PKGROOT)

oddcomm: $(SUBSYSTEMS) $(MODULES) $(CORE) $(PKGDIR)/lib/persist.a src/main/*.go
	$(GOCMD) -o oddcomm.$(O) $(wildcard src/main/*.go)
	rm -f oddcomm
	$(LDCMD) -o oddcomm oddcomm.$(O)
	rm -f oddcomm.$(O)

$(PKGDIR)/modules/%.a: $(SUBSYSTEMS) $(CORE) modules/%/*.go
	mkdir -p $(PKGDIR)/modules/$*
	rmdir $(PKGDIR)/modules/$*
	$(GOCMD) -o $(PKGDIR)/modules/$*.$(O) $(wildcard modules/$*/*.go)
	rm -f $(PKGDIR)/modules/$*.a
	$(GOPACK) grc $(PKGDIR)/modules/$*.a $(PKGDIR)/modules/$*.$(O)
	rm -f $(PKGDIR)/modules/$*.$(O)

$(PKGDIR)/src/ts6.a: $(CORE) $(PKGDIR)/lib/irc.a src/ts6/*.go
	mkdir -p $(PKGDIR)/src
	$(GOCMD) -o $(PKGDIR)/src/ts6.$(O) $(wildcard src/ts6/*.go)
	rm -f $(PKGDIR)/src/ts6.a
	$(GOPACK) grc $(PKGDIR)/src/ts6.a $(PKGDIR)/src/ts6.$(O)
	rm -f $(PKGDIR)/src/ts6.$(O)

$(PKGDIR)/src/client.a: $(CORE) $(PKGDIR)/lib/irc.a $(PKGDIR)/lib/perm.a src/client/*.go
	mkdir -p $(PKGDIR)/src
	$(GOCMD) -o $(PKGDIR)/src/client.$(O) $(wildcard src/client/*.go)
	rm -f $(PKGDIR)/src/client.a
	$(GOPACK) grc $(PKGDIR)/src/client.a $(PKGDIR)/src/client.$(O)
	rm -f $(PKGDIR)/src/client.$(O)

$(PKGDIR)/lib/%.a: $(CORE) lib/%/*.go
	mkdir -p $(PKGDIR)/lib
	$(GOCMD) -o $(PKGDIR)/lib/$*.$(O) $(wildcard lib/$*/*.go)
	rm -f $(PKGDIR)/lib/$*.a
	$(GOPACK) grc $(PKGDIR)/lib/$*.a $(PKGDIR)/lib/$*.$(O)
	rm -f $(PKGDIR)/lib/$*.$(O)

$(PKGDIR)/src/core.a: src/core/*.go $(PKGDIR)/src/core/store.a $(PKGDIR)/src/core/logic.a
	mkdir -p $(PKGDIR)/src
	$(GOCMD) -o $(PKGDIR)/src/core.$(O) $(wildcard src/core/*.go)
	rm -f $(PKGDIR)/src/core.a
	$(GOPACK) grc $(PKGDIR)/src/core.a $(PKGDIR)/src/core.$(O)
	rm -f $(PKGDIR)/src/core.$(O)

$(PKGDIR)/src/core/logic.a: src/core/logic/*.go $(PKGDIR)/src/core/connect.a $(PKGDIR)/src/core/store.a
	mkdir -p $(PKGDIR)/src/core
	$(GOCMD) -o $(PKGDIR)/src/core/logic.$(O) $(wildcard src/core/logic/*.go)
	rm -f $(PKGDIR)/src/core/logic.a
	$(GOPACK) grc $(PKGDIR)/src/core/logic.a $(PKGDIR)/src/core/logic.$(O)
	rm -f $(PKGDIR)/src/core/logic.$(O)

$(PKGDIR)/src/core/connect.a: src/core/connect/*.go $(PKGDIR)/src/core/connect/mmn.a
	mkdir -p $(PKGDIR)/src/core
	$(GOCMD) -o $(PKGDIR)/src/core/connect.$(O) $(wildcard src/core/connect/*.go)
	rm -f $(PKGDIR)/src/core/connect.a
	$(GOPACK) grc $(PKGDIR)/src/core/connect.a $(PKGDIR)/src/core/connect.$(O)
	rm -f $(PKGDIR)/src/core/connect.$(O)

$(PKGDIR)/src/core/connect/mmn.a: src/core/connect/mmn/*.go src/core/connect/mmn/mmn.pb.go
	mkdir -p $(PKGDIR)/src/core/connect/mmn
	$(GOCMD) -o $(PKGDIR)/src/core/connect/mmn.$(O) $(wildcard src/core/connect/mmn/*.go)
	rm -f $(PKGDIR)/src/core/connect/mmn.a
	$(GOPACK) grc $(PKGDIR)/src/core/connect/mmn.a $(PKGDIR)/src/core/connect/mmn.$(O)
	rm -f $(PKGDIR)/src/core/connect/mmn.$(O)

$(PKGDIR)/src/core/store.a: src/core/store/*.go $(PKGDIR)/lib/trie.a
	mkdir -p $(PKGDIR)/src/core
	$(GOCMD) -o $(PKGDIR)/src/core/store.$(O) $(wildcard src/core/store/*.go)
	rm -f $(PKGDIR)/src/core/store.a
	$(GOPACK) grc $(PKGDIR)/src/core/store.a $(PKGDIR)/src/core/store.$(O)
	rm -f $(PKGDIR)/src/core/store.$(O)

$(PKGDIR)/lib/trie.a: $(PKGDIR)/lib/cas.a lib/trie/main.go lib/trie/base.go
	cp lib/trie/base.go lib/trie/string.go
	sed -i 's/interface{}(nil)/""/g' lib/trie/string.go
	sed -i "s/interface{}/string/g" lib/trie/string.go
	sed -i "s/Trie/StringTrie/g" lib/trie/string.go
	sed -i "s/node/stringNode/g" lib/trie/string.go
	sed -i "s/Iterator/StringIterator/g" lib/trie/string.go
	cp lib/trie/base.go lib/trie/pointer.go
	sed -i 's/package trie/package trie\nimport "unsafe"/g' lib/trie/pointer.go
	sed -i "s/interface{}/unsafe.Pointer/g" lib/trie/pointer.go
	sed -i "s/Trie/PointerTrie/g" lib/trie/pointer.go
	sed -i "s/node/pointerNode/g" lib/trie/pointer.go
	sed -i "s/Iterator/PointerIterator/g" lib/trie/pointer.go
	mkdir -p $(PKGDIR)/lib
	$(GOCMD) -o $(PKGDIR)/lib/trie.$(O) $(wildcard lib/trie/*.go)
	rm -f $(PKGDIR)/lib/trie.a
	$(GOPACK) grc $(PKGDIR)/lib/trie.a $(PKGDIR)/lib/trie.$(O)
	rm -f $(PKGDIR)/lib/trie.$(O)

$(PKGDIR)/lib/cas.a: lib/cas/*.s lib/cas/*.go
	mkdir -p $(PKGDIR)/lib
	$(GOASM) -o $(PKGDIR)/lib/cas_asm.$(O) $(wildcard lib/cas/asm_$(GOARCH)*.s)
	$(GOCMD) -o $(PKGDIR)/lib/cas.$(O) $(wildcard lib/cas/*.go)
	rm -f $(PKGDIR)/lib/cas.a
	$(GOPACK) grc $(PKGDIR)/lib/cas.a $(PKGDIR)/lib/cas.$(O) $(PKGDIR)/lib/cas_asm.$(O)
	rm -f $(PKGDIR)/lib/cas.$(O) $(PKGDIR)/lib/cas_asm.$(O)


include $(GOROOT)/src/pkg/goprotobuf.googlecode.com/hg/Make.protobuf
