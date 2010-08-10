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
		if c := GetClient(u); c != nil {
			if oldvalue == "" {
				u.PermitRegistration()
			}
		}
	}, true)

	core.HookUserDataChanges(func(u *core.User, c *core.DataChange, old *core.OldData) {
		if cli := GetClient(u); c != nil {
			modeline := UserModes.ParseChanges(u, c, old)
			if modeline != "" {
				cli.WriteFrom(u, "MODE", modeline)
			}
		}
	}, false)

	core.HookUserRegister(func(u *core.User) {
		if c := GetClient(u); c != nil {
			c.WriteFrom(nil, "001", ":Welcome to IRC")
		}
	})

	core.HookUserPM("", func(source, target *core.User, message []byte) {
		if c := GetClient(target); c != nil {
			c.WriteFrom(source, "PRIVMSG", ":%s", message)
		}
	})

	core.HookUserPM("reply",
	                func(source, target *core.User, message []byte) {
		if c := GetClient(target); c != nil {
			c.WriteFrom(source, "NOTICE", ":%s", message)
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
