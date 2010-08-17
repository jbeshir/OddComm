DESTDIR=.

all:	clean install

install: build
	install src/main/oddircd $(DESTDIR)

build: 
	cd src/core && make install
	cd src/perm && make install
	cd src/irc && make install
	cd src/client && make install
	cd modules/dev/catserv && make install
	cd modules/dev/horde && make install
	cd modules/dev/tmmode && make install
	cd modules/irc/botmode && make install
	cd src/main && make

clean:
	cd src/core && make clean
	cd src/perm && make clean
	cd src/irc && make clean
	cd src/client && make clean
	cd modules/dev/catserv && make clean
	cd modules/dev/horde && make clean
	cd modules/dev/tmmode && make clean
	cd modules/irc/botmode && make clean
	cd src/main && make clean

nuke:
	cd src/core && make nuke
	cd src/perm && make nuke
	cd src/irc && make nuke
	cd src/client && make nuke
	cd modules/dev/catserv && make nuke
	cd modules/dev/horde && make nuke
	cd modules/dev/tmmode && make nuke
	cd modules/irc/botmode && make nuke
	cd src/main && make nuke
