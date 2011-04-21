package ts6

import "fmt"

import "oddcomm/src/core"


func init() {
	core.HookUserRegister(func(_ interface{}, u *core.User) {
		// Don't introduce them if they came *from* this module.
		if u.Owner() != me {
			all(func(l *local) { send_uid(l, u) })
		}
	})

	core.HookUserNickChange(func(origin interface{}, u *core.User, _, newnick string, ts int64) {
		s, _ := origin.(*server)

		all(func(l *local) {
			l.propagate_user(s, u, fmt.Sprintf(":%s NICK %s %d\r\n",
				u.ID(),	newnick, ts))
		})
	}, false)

	core.HookUserDelete(func(origin interface{}, source, u *core.User, msg string) {
		s, _ := origin.(*server)

		if source == u {
			all(func(l *local) {
				l.propagate_user(s, u,
					fmt.Sprintf(":%s QUIT :Quit: %s\r\n", u.ID(),
					msg))
			})
		} else if source != nil {
			all(func(l *local) {
				l.propagate_user(s, u,
					fmt.Sprintf(":%s KILL %s :%s\r\n", source.ID(),
					u.ID(), msg))
			})
		} else {
			all(func(l *local) {
				l.propagate_user(s, u, fmt.Sprintf(":%s QUIT :%s\r\n",
					u.ID(), msg))
			})
		}
	}, false)

	core.HookUserMessage("", func(origin interface{}, source, target *core.User, msg []byte) {
		s, _ := origin.(*server)

		if target.Owner() != me {
			return
		}

		dest := target.Owndata().(*server)
		dest.local.propagate_msg(s, target,
			fmt.Sprintf(":%s PRIVMSG %s :%s\r\n", from(source),
				target.ID(), msg))
	})

	core.HookUserMessage("noreply", func(origin interface{}, source, target *core.User, msg []byte) {
		s, _ := origin.(*server)

		if target.Owner() != me {
			return
		}

		dest := target.Owndata().(*server)
		dest.local.propagate_msg(s, target,
			fmt.Sprintf(":%s NOTICE %s :%s\r\n", from(source), target.ID(),
				msg))
	})
}
