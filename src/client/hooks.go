package client

import "fmt"

import "oddcomm/src/core"


func init() {
	core.HookUserNickChange(func(u *core.User, oldnick, newnick string) {
		sent := make(map[*Client]bool)

		if c := GetClient(u); c != nil {
			fmt.Fprintf(c, ":%s!%s@%s NICK %s\r\n", oldnick,
					u.Data("ident"), u.Data("hostname"),
					u.Nick())
			sent[c] = true
		}

		// Send the nick change to every user on every channel the
		// user is on.
		for ch := u.Channels(); ch != nil; ch = ch.UserNext() {
			for m := ch.Channel().Users(); m != nil; m = m.ChanNext() {
				c := GetClient(m.User())
				if c == nil || sent[c] {
					continue
				}
				fmt.Fprintf(c, ":%s!%s@%s NICK %s\r\n",
				            oldnick, u.Data("ident"),
				            u.Data("hostname"), u.Nick())
				sent[c] = true
			}
		}
	}, false)

	core.RegistrationHold("oddcomm/src/client")
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
				cli.WriteTo(source, "MODE", modeline)
			}
		}
	}, false)

	core.HookUserRegister(func(u *core.User) {
		if c := GetClient(u); c != nil {
			c.WriteTo(nil, "001", ":Welcome to IRC")
			modeline := UserModes.GetModes(u)
			c.WriteTo(u, "MODE", "+%s", modeline)
		}
	})

	core.HookUserMessage("", func(source, target *core.User,
			message []byte) {
		if c := GetClient(target); c != nil {
			c.WriteTo(source, "PRIVMSG", ":%s", message)
		}
	})

	core.HookUserMessage("noreply",
	                func(source, target *core.User, message []byte) {
		if c := GetClient(target); c != nil {
			c.WriteTo(source, "NOTICE", ":%s", message)
		}
	})

	core.HookChanUserJoin("", func(u *core.User, ch *core.Channel) {

		// Send the JOIN to all clients in the same channel.
		for m := ch.Users(); m != nil; m = m.ChanNext() {
			if c := GetClient(m.User()); c != nil {
				c.WriteFrom(u, "JOIN #%s", ch.Name())
			}
		}

		// If it's our client that joined...
		if c := GetClient(u); c != nil {

			// Send them NAMES.
			var params [1][]byte
			params[0] = []byte(ch.Name())
			cmdNames(u, c, params[0:])

			// Send them the topic.
			if topic, setby, setat := ch.GetTopic(); topic != "" {
				c.WriteTo(nil, "332", "#%s :%s", ch.Name(),
				          topic)
				c.WriteTo(nil, "333", "#%s %s %s", ch.Name(),
				          setby, setat)
			}
		}
	})

	core.HookChanDataChange("", "topic",
	                        func(source *core.User, ch *core.Channel, oldvalue, newvalue string) {
		for m := ch.Users(); m != nil; m = m.ChanNext() {
			if c := GetClient(m.User()); c != nil {
				c.WriteFrom(source, "TOPIC #%s :%s", ch.Name(), newvalue)
			}
		}
	})

	core.HookChanDataChanges("", func(source *core.User, ch *core.Channel, c *core.DataChange, old *core.OldData) {
		modeline := ChanModes.ParseChanges(ch, c, old)
		if modeline == "" {
			return
		}
		for m := ch.Users(); m != nil; m = m.ChanNext() {
			c := GetClient(m.User())
			if c == nil {
				continue
			}
			c.WriteFrom(source, "MODE #%s %s", ch.Name(), modeline)
		}
	})

	core.HookChanUserRemove("", func(source, u *core.User,
	                                 ch *core.Channel, message string) {
		// If the user doesn't exist anymore (quit, for example),
		// don't bother to show their removal, we'll already have
		// shown the quit.
		if core.GetUser(u.ID()) != u {
			return
		}

		// Send a PART or KICK to the user themselves.
		if c := GetClient(u); c != nil {
			if source == u {
				c.WriteFrom(u, "PART #%s :%s", ch.Name(),
				            message)
			} else {
				c.WriteFrom(source, "KICK #%s %s :%s",
				            ch.Name(), u.Nick(), message)
			}
		}

		// Send a PART or KICK to everyone in the user's channel.
		for m := ch.Users(); m != nil; m = m.ChanNext() {
			if c := GetClient(m.User()); c != nil {
				if source == u {
					c.WriteFrom(u, "PART #%s :%s",
					            ch.Name(), message)
				} else {
					c.WriteFrom(source, "KICK #%s %s :%s",
					            ch.Name(), u.Nick(),
					            message)
				}
			}
		}
	})

	core.HookChanMessage("", "", func(source *core.User, ch *core.Channel,
	                     message []byte) {
		for m := ch.Users(); m != nil; m = m.ChanNext() {
			if m.User() == source {
				continue
			}
			if c := GetClient(m.User()); c != nil {
				c.WriteFrom(source, "PRIVMSG #%s :%s",
				            ch.Name(), message)
			}
		}
	})

	core.HookChanMessage("", "noreply",
	                     func(source *core.User, ch *core.Channel,
			     message []byte) {
		for m := ch.Users(); m != nil; m = m.ChanNext() {
			if m.User() == source {
				continue
			}
			if c := GetClient(m.User()); c != nil {
				c.WriteFrom(source, "NOTICE #%s :%s",
				            ch.Name(), message)
			}
		}
	})

	core.HookUserDelete(func(u *core.User, message string) {
		sent := make(map[*Client]bool)

		// If this is our client, delete them.
		if c := GetClient(u); c != nil {
			makeRequest(c, func() {
				c.delete(message)
			})		
			sent[c] = true
		}

		// Send the quit to every user on every channel the user is on.
		for ch := u.Channels(); ch != nil; ch = ch.UserNext() {
			for m := ch.Channel().Users(); m != nil; m = m.ChanNext() {
				c := GetClient(m.User())
				if c == nil || sent[c] {
					continue
				}
				c.WriteFrom(u, "QUIT :%s", message)
				sent[c] = true
			}
		}
	}, true)
}
