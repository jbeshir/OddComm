package client

import "fmt"

import "oddircd/src/core"


func init() {
	core.HookUserNickChange(func(u *core.User, oldnick, newnick string) {
		if !u.Registered() {
			return
		}

		if c := GetClient(u); c != nil {
			fmt.Fprintf(c, ":%s!%s@%s NICK %s\r\n", oldnick,
					u.Data("ident"), u.Data("hostname"),
					u.Nick())
		}
	}, true)

	core.RegistrationHold("oddircd/src/client")
	core.HookUserDataChange("ident",
	                        func(u *core.User, oldvalue, newvalue string) {
		if c := clients_by_user[u]; c != nil {
			if oldvalue == "" {
				u.PermitRegistration()
			}
		}
	}, true)

	core.HookUserRegister(func(u *core.User) {
		if c := GetClient(u); c != nil {
			fmt.Fprintf(c, ":%s 001 %s :Welcome to IRC\r\n", "Server.name", u.Nick())
		}
	})

	core.HookUserPM("", func(source, target *core.User, message string) {
		if c := GetClient(target); c != nil {
			if source != nil {
				fmt.Fprintf(c, ":%s!%s@%s PRIVMSG %s :%s\r\n",
				            source.Nick(), source.Data("ident"),
				            source.Data("hostname"),
				            target.Nick(), message)
			} else {
				fmt.Fprintf(c,
				            ":Server.name PRIVMSG %s :%s\r\n",
				            source.Nick(), source.Data("ident"),
				            source.Data("hostname"),
				            target.Nick(), message)
			}
		}
	})

	core.HookUserRemoved(func(u *core.User, message string) {
		if c := GetClient(u); c != nil {
			makeRequest(c, func() {
				c.remove(message)
			})		
		}
	}, true)
}
