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

	core.HookUserMessage("", func(pkg string, source, target *core.User, msg []byte) {

		if pkg == me || target.Owner() != me {
			return
		}

		s := target.Owndata().(*server)
		fmt.Fprintf(s.local, ":%s PRIVMSG %s :%s\n", source.ID(),
			target.ID(), msg)
	})

	core.HookUserMessage("noreply", func(pkg string, source, target *core.User, msg []byte) {

		if pkg == me || target.Owner() != me {
			return
		}

		s := target.Owndata().(*server)
		fmt.Fprintf(s.local, ":%s NOTICE %s :%s\n", source.ID(),
			target.ID(), msg)
	})
}
