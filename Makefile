DESTDIR=.

all:	clean install

install: build
	install src/main/oddircd $(DESTDIR)

build: 
	cd src/core && make install
	cd src/irc && make install
	cd src/client && make install
	cd src/main && make clean && make

clean:
	cd src/core && make clean
	cd src/irc && make clean
	cd src/client && make clean
	cd src/main && make clean

nuke:
	cd src/core && make nuke
	cd src/irc && make nuke
	cd src/client && make nuke
	cd src/main && make nuke
