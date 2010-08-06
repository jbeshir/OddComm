package client

import "fmt"

import "oddircd/core"


func init() {
	core.HookNickChange(func(u *core.User, oldnick string) {
		if oldnick == "" || !u.Registered() {
			return
		}

		if c := clients_by_user[u]; c != nil {
			fmt.Fprintf(c, "NICK %s!%s@%s %s\r\n", oldnick,
					u.Data("ident"), u.Data("hostname"),
					u.Nick())
		}
	}, true)

	core.IncrementRegcount()
	core.HookDataChange("ident", func(u *core.User, oldvalue string) {
		if c := clients_by_user[u]; c != nil {
			if oldvalue == "" {
				u.DecrementRegcount()
			}
		}
	}, true)

	core.HookRemoved(func(u *core.User, message string) {
		if c := clients_by_user[u]; c != nil {
			makeRequest(c, func() {
				c.remove(message)
			})		
		}
	}, true)

	core.HookRegistered(func(u *core.User) {
		if c := clients_by_user[u]; c != nil {
			fmt.Fprintf(c, ":%s 001 %s :Welcome to IRC\r\n", "Server.name", u.Nick())
		}
	})
}
