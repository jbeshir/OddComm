package ts6

import "fmt"

import "oddcomm/src/core"


func init() {
	core.HookUserRegister(func(_ string, u *core.User) {
		// Don't introduce them if they came *from* this module.
		if u.Owner() != me {
			send_uid(all, u)
		}
	})

	core.HookUserDelete(func(pkg string, source, u *core.User, msg string) {
		if pkg == me {
			return
		}

		fmt.Fprintf(all, ":%s QUIT %s\r\n", u.ID(), msg)
	}, false)

	core.HookUserMessage("", func(pkg string, source, target *core.User, msg []byte) {

		if pkg == me || target.Owner() != me {
			return
		}

		s := target.Owndata().(*server)
		fmt.Fprintf(s.local, ":%s PRIVMSG %s :%s\r\n", source.ID(),
			target.ID(), msg)
	})

	core.HookUserMessage("noreply", func(pkg string, source, target *core.User, msg []byte) {

		if pkg == me || target.Owner() != me {
			return
		}

		s := target.Owndata().(*server)
		fmt.Fprintf(s.local, ":%s NOTICE %s :%s\r\n", source.ID(),
			target.ID(), msg)
	})
}
