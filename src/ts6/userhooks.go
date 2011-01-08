package ts6

import "fmt"

import "oddcomm/src/core"


func init() {
	core.HookUserRegister(func(u *core.User) {
		// Don't introduce them if they came *from* this module.
		if u.Owner() != "oddcomm/src/ts6" {
			send_uid(all, u)
		}
	})

	core.HookUserMessage("", func(source, target *core.User, msg []byte) {
		if target.Owner() == "oddcomm/src/ts6" {
			s, ok := target.Owndata().(*server)
			if !ok {
				// !?!
				return
			}

			fmt.Fprintf(s.local, ":%s PRIVMSG %s :%s\n",
					source.ID(), target.ID(), msg)
		}
	})

	core.HookUserMessage("noreply", func(source, target *core.User,
			msg []byte) {
		if target.Owner() == "oddcomm/src/ts6" {
			s, ok := target.Owndata().(*server)
			if !ok {
				// !?!
				return
			}

			fmt.Fprintf(s.local, ":%s NOTICE %s :%s\n",
					source.ID(), target.ID(), msg)
		}
	})
}
