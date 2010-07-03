DESTDIR=.

all:	install

install:
	cd src/core && make install
	cd src/irc && make install
	cd src/client && make install
	cd src/main && make clean && make
	install src/main/oddircd $(DESTDIR)

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
