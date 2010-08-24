DESTDIR=.

all:	clean install

install: build
	install src/main/oddircd $(DESTDIR)

build: 
	cd src/core && make install
	cd src/perm && make install
	cd src/irc && make install
	cd src/client && make install
	cd modules/client/botmode && make install
	cd modules/client/extbans && make install
	cd modules/client/login && make install
	cd modules/client/ochanctrl && make install
	cd modules/client/opermode && make install
	cd modules/client/opflags && make install
	cd modules/oper/account && make install
	cd modules/oper/pmoverride && make install
	cd modules/dev/catserv && make install
	cd modules/dev/horde && make install
	cd modules/dev/tmmode && make install
	cd src/main && make

clean:
	cd src/core && make clean
	cd src/perm && make clean
	cd src/irc && make clean
	cd src/client && make clean
	cd modules/client/botmode && make clean
	cd modules/client/extbans && make clean
	cd modules/client/login && make clean
	cd modules/client/ochanctrl && make clean
	cd modules/client/opermode && make clean
	cd modules/client/opflags && make clean
	cd modules/oper/account && make clean
	cd modules/oper/pmoverride && make clean
	cd modules/dev/catserv && make clean
	cd modules/dev/horde && make clean
	cd modules/dev/tmmode && make clean
	cd src/main && make clean

nuke:
	cd src/core && make nuke
	cd src/perm && make nuke
	cd src/irc && make nuke
	cd src/client && make nuke
	cd modules/client/botmode && make nuke
	cd modules/client/extbans && make nuke
	cd modules/client/login && make nuke
	cd modules/client/ochanctrl && make nuke
	cd modules/client/opermode && make nuke
	cd modules/client/opflags && make nuke
	cd modules/oper/account && make nuke
	cd modules/oper/pmoverride && make nuke
	cd modules/dev/catserv && make nuke
	cd modules/dev/horde && make nuke
	cd modules/dev/tmmode && make nuke
	cd src/main && make nuke
