package client

import "fmt"

import "oddcomm/src/core"


func init() {
	core.HookUserNickChange(func(_ interface{}, u *core.User, oldnick, newnick string, _ int64) {
		sent := make(map[*Client]bool)

		// Send the nick change to every user on a common channel.
		for ch := u.Channels(); ch != nil; ch = ch.UserNext() {
			for m := ch.Channel().Users(); m != nil; m = m.ChanNext() {
				chu := m.User()

				c := GetClient(chu)
				if c == nil || sent[c] {
					continue
				}

				fmt.Fprintf(c, ":%s!%s@%s NICK %s\r\n", oldnick,
					u.Data("ident"), u.Data("hostname"),
					u.Nick())

				sent[c] = true
			}
		}

		c := GetClient(u)
		if c == nil || sent[c] {
			return
		}

		fmt.Fprintf(c, ":%s!%s@%s NICK %s\r\n", oldnick,
			u.Data("ident"), u.Data("hostname"), u.Nick())
	},
		false)

	core.RegistrationHold(me)
	core.HookUserDataChange("ident", func(_ interface{}, source, target *core.User, oldvalue, newvalue string) {
		if target.Owner() != me {
			return
		}

		if oldvalue == "" {
			target.PermitRegistration(me)
		}
	},
		true)

	core.HookUserDataChanges(func(_ interface{}, source, target *core.User, c []core.DataChange, old []string) {
		cli := GetClient(target)
		if cli == nil {
			return
		}

		modeline := UserModes.ParseChanges(target, c, old)
		if modeline != "" {
			cli.SendLineTo(source, "MODE", modeline)
		}
	},
		false)

	core.HookUserRegister(func(_ interface{}, u *core.User) {
		c := GetClient(u)
		if c == nil {
			return
		}

		c.SendLineTo(nil, "001", ":Welcome to the %s IRC Network %s!%s@%s", "Testnet", u.Nick(), u.GetIdent(), u.GetHostname())
		c.SendLineTo(nil, "002", "Your host is %s, running version OddComm-%s", core.Global.Data("name"), core.Version)
		c.SendLineTo(nil, "004", "%s OddComm-%s %s%s%s %s %s%s%s", core.Global.Data("name"), core.Version, UserModes.AllSimple(), UserModes.AllParametered(), UserModes.AllList(), ChanModes.AllSimple(), ChanModes.AllParametered(), ChanModes.AllList(), ChanModes.AllMembership())
		c.SendLineTo(nil, "005", "%s :are supported by this server", supportLine)
		c.SendLineTo(nil, "005", "%s :your unique ID", u.ID())
		modeline := UserModes.GetModes(u)
		c.SendLineTo(u, "MODE", "+%s", modeline)
	})

	core.HookUserMessage("", func(_ interface{}, source, target *core.User, message []byte) {
		c := GetClient(target)
		if c == nil {
			return
		}

		c.SendLineTo(source, "PRIVMSG", ":%s", message)
	})

	core.HookUserMessage("noreply",
		func(_ interface{}, source, target *core.User, message []byte) {
			c := GetClient(target)
			if c == nil {
				return
			}

			c.SendLineTo(source, "NOTICE", ":%s", message)
		})

	core.HookUserMessage("invite",
		func(_ interface{}, source, target *core.User, message []byte) {
			c := GetClient(target)
			if c == nil {
				return
			}

			c.SendLineTo(source, "INVITE", ":#%s", message)
		})

	core.HookChanUserJoin("", func(origin interface{}, ch *core.Channel, users []*core.User) {
		pkg, _ := origin.(string)

		// Send the JOINs to all clients in the same channel.
		for m := ch.Users(); m != nil; m = m.ChanNext() {
			chu := m.User()
			c := GetClient(chu)
			if c == nil {
				continue
			}

			for _, u := range users {
				if chu == u && pkg == me {
					continue
				}
				c.SendFrom(u, "JOIN #%s", ch.Name())
			}
		}

		// If we did the join, that's it.
		if pkg == me {
			return
		}

		// Otherwise, if this is one of our clients,
		// we need to send them info.
		for _, u := range users {
			c := GetClient(u)
			if c == nil {
				continue
			}

			// Done concurrently, since sending NAMES can block.
			go func() {
				// Send them NAMES.
				cmdNames(c, [][]byte{[]byte(ch.Name())})

				// Send them the topic.
				if topic, setby, setat := ch.GetTopic(); topic != "" {
					c.SendLineTo(nil, "332", "#%s :%s", ch.Name(),
						topic)
					c.SendLineTo(nil, "333", "#%s %s %s",
						ch.Name(), setby, setat)
				}
			}()
		}
	})

	core.HookChanDataChange("", "topic", func(_ interface{}, source *core.User, ch *core.Channel, _, newvalue string) {
		for m := ch.Users(); m != nil; m = m.ChanNext() {
			c := GetClient(m.User())
			if c == nil {
				continue
			}

			c.SendFrom(source, "TOPIC #%s :%s", ch.Name(), newvalue)
		}
	})

	core.HookChanDataChanges("", func(_ interface{}, source *core.User, ch *core.Channel, c []core.DataChange, old []string) {
		modeline := ChanModes.ParseChanges(ch, c, old)
		if modeline == "" {
			return
		}
		for m := ch.Users(); m != nil; m = m.ChanNext() {
			c := GetClient(m.User())
			if c == nil {
				continue
			}
			c.SendFrom(source, "MODE #%s %s", ch.Name(), modeline)
		}
	})

	core.HookChanUserRemove("", func(_ interface{}, source, u *core.User, ch *core.Channel, message string) {
		// If the user isn't registered anymore, they quit.
		// Don't bother to show their part.
		if !u.Registered() {
			return
		}

		// Send a PART or KICK to the user themselves.
		if c := GetClient(u); c != nil {
			if source == u {
				c.SendFrom(u, "PART #%s :%s", ch.Name(),
					message)
			} else {
				c.SendFrom(source, "KICK #%s %s :%s",
					ch.Name(), u.Nick(), message)
			}
		}

		// Send a PART or KICK to everyone in the user's channel.
		for m := ch.Users(); m != nil; m = m.ChanNext() {
			if c := GetClient(m.User()); c != nil {
				if source == u {
					c.SendFrom(u, "PART #%s :%s",
						ch.Name(), message)
				} else {
					c.SendFrom(source, "KICK #%s %s :%s",
						ch.Name(), u.Nick(),
						message)
				}
			}
		}
	})

	core.HookChanMessage("", "", func(_ interface{}, source *core.User, ch *core.Channel, message []byte) {
		for m := ch.Users(); m != nil; m = m.ChanNext() {
			if m.User() == source {
				continue
			}
			if c := GetClient(m.User()); c != nil {
				c.SendFrom(source, "PRIVMSG #%s :%s",
					ch.Name(), message)
			}
		}
	})

	core.HookChanMessage("", "noreply", func(_ interface{}, source *core.User, ch *core.Channel, message []byte) {
		for m := ch.Users(); m != nil; m = m.ChanNext() {
			if m.User() == source {
				continue
			}
			if c := GetClient(m.User()); c != nil {
				c.SendFrom(source, "NOTICE #%s :%s",
					ch.Name(), message)
			}
		}
	})

	core.HookChanMessage("", "invite", func(_ interface{}, source *core.User, ch *core.Channel, message []byte) {
		for m := ch.Users(); m != nil; m = m.ChanNext() {
			if c := GetClient(m.User()); c != nil {
				c.SendFrom(nil, "NOTICE #%s :*** INVITE: %s invited %s into the channel.", ch.Name(), source.Nick(), message)
			}
		}
	})

	core.HookUserDelete(func(_ interface{}, source, u *core.User, message string) {
		sent := make(map[*Client]bool)

		// Send a KILL message to the user, if they were deleted by
		/// another user and are our client.
		if c := GetClient(u); c != nil {
			if source != nil && source != u {
				c.SendLineTo(source, "KILL", "%s (%s)", source.Nick(), message)
			}
		}

		// Add text to the message to indicate its source.
		if source == u {
			message = "Quit: " + message
		} else if source != nil {
			message = "Killed by " + source.Nick() + ": " + message
		}

		// If this is our client, delete them.
		if c := GetClient(u); c != nil {
			c.mutex.Lock()
			c.delete(message)
			c.mutex.Unlock()
			sent[c] = true
		}

		// Send the quit to every user on every channel the user is on.
		for ch := u.Channels(); ch != nil; ch = ch.UserNext() {
			for m := ch.Channel().Users(); m != nil; m = m.ChanNext() {
				c := GetClient(m.User())
				if c == nil || sent[c] {
					continue
				}
				c.SendFrom(u, "QUIT :%s", message)
				sent[c] = true
			}
		}
	},
		true)
}
