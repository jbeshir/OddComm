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
	                        func(source, target *core.User, oldvalue, newvalue string) {
		if c := GetClient(target); c != nil {
			if oldvalue == "" {
				target.PermitRegistration()
			}
		}
	}, true)

	core.HookUserDataChanges(func(source, target *core.User, c *core.DataChange, old *core.OldData) {
		if cli := GetClient(target); c != nil {
			modeline := UserModes.ParseChanges(target, c, old)
			if modeline != "" {
				cli.WriteFrom(source, "MODE", modeline)
			}
		}
	}, false)

	core.HookUserRegister(func(u *core.User) {
		if c := GetClient(u); c != nil {
			c.WriteFrom(nil, "001", ":Welcome to IRC")
			modeline := UserModes.GetModes(u)
			c.WriteFrom(u, "MODE", "+%s", modeline)
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

	core.HookUserDelete(func(u *core.User, message string) {
		if c := GetClient(u); c != nil {
			makeRequest(c, func() {
				c.delete(message)
			})		
		}
	}, true)
}
