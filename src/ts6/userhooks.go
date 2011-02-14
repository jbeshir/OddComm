package ts6

import "oddcomm/src/core"
import "oddcomm/lib/irc"


func init() {
	core.HookUserRegister(func(_ string, u *core.User) {
		// Don't introduce them if they came *from* this module.
		if u.Owner() != me {
			all(func(l *local) { send_uid(l, u) })
		}
	})

	core.HookUserNickChange(func(pkg string, u *core.User, _, newnick string, ts int64) {
		if pkg == me {
			return
		}

		all(func(l *local) {
			irc.SendFrom(l, from(u), "NICK %s %d", newnick, ts)
		})
	}, false)

	core.HookUserDelete(func(pkg string, source, u *core.User, msg string) {
		if pkg == me {
			return
		}

		if source == u {
			all(func(l *local) {
				irc.SendFrom(l, from(u), "QUIT :Quit: %s", msg)
			})
		} else if source != nil {
			all(func(l *local) {
				irc.SendFrom(l, from(source), "KILL %s :%s", u.ID(),
					msg)
			})
		} else {
			all(func(l *local) {
				irc.SendFrom(l, from(u), "QUIT :%s", msg)
			})
		}
	}, false)

	core.HookUserMessage("", func(pkg string, source, target *core.User, msg []byte) {

		if pkg == me || target.Owner() != me {
			return
		}

		s := target.Owndata().(*server)
		irc.SendLine(s.local, from(source), from(target), "PRIVMSG", ":%s",
			msg)
	})

	core.HookUserMessage("noreply", func(pkg string, source, target *core.User, msg []byte) {

		if pkg == me || target.Owner() != me {
			return
		}

		s := target.Owndata().(*server)
		irc.SendLine(s.local, from(source), from(target), "NOTICE", ":%s", msg)
	})
}
