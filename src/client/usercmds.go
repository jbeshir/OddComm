package client

import "io"
import "strings"

import "oddircd/src/core"
import "oddircd/src/perm"
import "oddircd/src/irc"


// Add core user commands.
func init() {
	var c *irc.Command

	c = new(irc.Command)
	c.Handler = cmdUser
	c.Minargs = 4
	c.Maxargs = 4
	c.Unregged = 2
	Commands.Add("USER", c)

	c = new(irc.Command)
	c.Handler = cmdNick
	c.Minargs = 1
	c.Maxargs = 1
	c.Unregged = 1
	Commands.Add("NICK", c)

	c = new(irc.Command)
	c.Handler = irc.CmdQuit
	c.Maxargs = 1
	c.Unregged = 1
	Commands.Add("QUIT", c)

	c = new(irc.Command)
	c.Handler = cmdPing
	c.Minargs = 1
	c.Maxargs = 1
	Commands.Add("PING", c)

	c = new(irc.Command)
	c.Handler = cmdWho
	c.Minargs = 1
	c.Maxargs = 1
	Commands.Add("WHO", c)

	c = new(irc.Command)
	c.Handler = cmdNames
	c.Minargs = 1
	c.Maxargs = 1
	Commands.Add("NAMES", c)

	c = new(irc.Command)
	c.Handler = cmdJoin
	c.Minargs = 1
	c.Maxargs = 1
	Commands.Add("JOIN", c)

	c = new(irc.Command)
	c.Handler = cmdPart
	c.Minargs = 1
	c.Maxargs = 1
	Commands.Add("PART", c)
	
	c = new(irc.Command)
	c.Handler = cmdKick
	c.Minargs = 2
	c.Maxargs = 3
	Commands.Add("KICK", c)
	
	c = new(irc.Command)
	c.Handler = cmdMode
	c.Minargs = 1
	c.Maxargs = 42
	Commands.Add("MODE", c)

	c = new(irc.Command)
	c.Handler = cmdTopic
	c.Minargs = 1
	c.Maxargs = 2
	Commands.Add("TOPIC", c)

	c = new(irc.Command)
	c.Handler = cmdPrivmsg
	c.Minargs = 2
	c.Maxargs = 2
	Commands.Add("PRIVMSG", c)

	c = new(irc.Command)
	c.Handler = cmdNotice
	c.Minargs = 2
	c.Maxargs = 2
	Commands.Add("NOTICE", c)
}

func cmdNick(u *core.User, w io.Writer, params [][]byte) {
	var nick = string(params[0])

	if nick == "0" {
		nick = u.ID()		
	}

	if ok, err := perm.CheckNick(u, nick); !ok {
		if c, ok := w.(*Client); ok {
			c.WriteTo(nil, "432", "%s :%s", nick, err)
		}
		return
	}

	if err := u.SetNick(nick); err != nil {
		if c, ok := w.(*Client); ok {
			c.WriteTo(nil, "433", "%s :%s", nick, err)
		}
	}
}

func cmdUser(u *core.User, w io.Writer, params [][]byte) {
	c := w.(*Client)
	if (u.Data("ident") != "") { return }

	ident := "~" + string(params[0])
	real := string(params[3])

	// Check that the ident and realname are valid.
	if num, err := perm.CheckUserDataPerm(u, u, "ident", ident); num <= -1e9 {
		c.WriteTo(nil, "461", "USER :%s", err)
		return
	}
	if num, err := perm.CheckUserDataPerm(u, u, "realname", real); num <= -1e9 {
		c.WriteTo(nil, "461", "USER :%s", err)
		return
	}
	
	u.SetData(nil, "ident", ident)
	u.SetData(nil, "real", real)
}

func cmdPing(u *core.User, w io.Writer, params [][]byte) {
	c := w.(*Client)
	c.WriteFrom(nil, "PONG %s :%s", "Server.name", params[0])
	
}

func cmdWho(u *core.User, w io.Writer, params [][]byte) {
	c := w.(*Client)
	channame := string(params[0])
	if channame[0] == '#' {
		channame = channame[1:]
	}

	if ch := core.FindChannel("", channame); ch != nil {
		for it := ch.Users(); it != nil; it = it.ChanNext() {
			user := it.User()
			var prefixes string
			if user.Data("away") == "" {
				prefixes += "H"
			} else {
				prefixes += "G"
			}
			if user.Data("op") != "" {
				prefixes += "*"
			}
			prefixes += ChanModes.GetPrefixes(it)
			c.WriteTo(nil, "352", "#%s %s %s %s %s %s :0 %s",
			          channame, user.GetIdent(),
			          user.GetHostname(), "Server.name",
			          user.Nick(), prefixes, user.Data("realname"))
		}
		c.WriteTo(nil, "315", "#%s :End of /WHO list.", channame)
	}
}

func cmdNames(u *core.User, w io.Writer, params [][]byte) {
	c := w.(*Client)
	channame := string(params[0])
	if channame[0] == '#' {
		channame = channame[1:]
	}

	if ch := core.FindChannel("", channame); ch != nil {
		var names string
		for it := ch.Users(); it != nil; it = it.ChanNext() {
			names += " " + ChanModes.GetPrefixes(it)
			names += it.User().Nick()
		}

		var myprefix string
		if m := ch.GetMember(u); m != nil {
			myprefix = ChanModes.GetPrefixes(m)
		}
		if myprefix == "" {
			myprefix = "="
		}
		c.WriteTo(nil, "353", "%s #%s :%s", myprefix, channame, names)
		c.WriteTo(nil, "366", "#%s :End of /NAMES list", channame)
	}
}

func cmdJoin(u *core.User, w io.Writer, params [][]byte) {
	chans := strings.Split(string(params[0]), ",", -1)
	for _, c := range chans {
		channame := c
		if channame[0] == '#' {
			channame = channame[1:]
		}
	
		core.GetChannel("", channame).Join(u)
	}
}

func cmdPart(u *core.User, w io.Writer, params [][]byte) {
	chans := strings.Split(string(params[0]), ",", -1)
	for _, c := range chans {
		channame := c
		if channame[0] == '#' {
			channame = channame[1:]
		}

		if ch := core.FindChannel("", channame); ch != nil {
			ch.Remove(u, u)
		}
	}
}

func cmdKick(u *core.User, w io.Writer, params [][]byte) {
	c := w.(*Client)

	var ch *core.Channel
	channame := string(params[0])
	if channame[0] == '#' {
		channame = channame[1:]
	}
	if ch = core.FindChannel("", channame); ch == nil {
		return
	}

	nicks := strings.Split(string(params[1]), ",", -1)
	for _, nick := range nicks {
		target := core.GetUserByNick(nick)
		if target == nil {
			continue
		}
		if ok, err := perm.CheckRemove(u, target, ch); ok {
			ch.Remove(u, target)
		} else {
			c.WriteTo(nil, "482", "#%s :%s", ch.Name(), err)
		}
	}
}

func cmdMode(u *core.User, w io.Writer, params [][]byte) {
	c := w.(*Client)

	// If we're viewing the modes of ourselves or a channel...
	if len(params) < 2 {
		if strings.ToUpper(u.Nick()) == strings.ToUpper(string(params[0])) {
			modeline := UserModes.GetModes(u)
			c.WriteTo(nil, "221", ":+%s", modeline)
			return
		}

		if params[0][0] == '#' {
			channame := string(params[0][1:])
			ch := core.FindChannel("", channame)
			if ch == nil {
				return
			}
			modeline := ChanModes.GetModes(ch)
			ts := ch.TS()
			c.WriteTo(nil, "324", "#%s +%s", ch.Name(), modeline)
			c.WriteTo(nil, "329", "#%s %d", ch.Name(), ts)
			return
		}
		return
	}

	// If we're listing list modes on a channel...
	if params[0][0] == '#' && params[1][0] != '+' && params[1][0] != '-' {
		channame := string(params[0][1:])
		ch := core.FindChannel("", channame)
		if ch == nil {
			return
		}

		var badmodes string
		for _, mode := range string(params[1]) {
			name := ChanModes.GetName(mode)
			if name == "" {
				badmodes += string(mode)
				continue
			}

			if ok, err := perm.CheckChanViewData(u, ch, name); !ok {
				c.WriteTo(nil, "482", "#%s :%s", ch.Name(), err)
				return
			}

			// Different, fixed numerics for different
			// modes. Stupid protocol.
			num := "941"; endnum := "940"
			switch mode {
			case 'b': num = "367"; endnum = "368"
			case 'e': num = "348"; endnum = "349"
			case 'I': num = "346"; endnum = "347"
			}

			ChanModes.ListMode(ch, int(mode), func(p, v string) {
				var setTime string = "0"
				var setBy string = "Server.name"
				words := strings.Fields(v)
				for _, word := range words {
					if len(word) > 6 && word[0:6] == "setat-" {
						setTime = word[6:]
						continue
					}
					if len(word) > 6 && word[0:6] == "setby-" {
						setBy = word[6:]
						continue
					}
				}

				c.WriteTo(nil, num, "#%s %s %s %s",
				          ch.Name(), p, setBy, setTime)
			})
			c.WriteTo(nil, endnum, "#%s :End of mode list.",
				  ch.Name())
		}
		if badmodes != "" {
			if badmodes != string(params[1]) {
				c.WriteTo(nil, "501", "Unknown list modes: %s", badmodes)
				return
			}
			// If ALL the mode characters were invalid, we let it
			// fall through and try to treat it as setting modes.
		} else {
			return
		}
	}

	// If we're setting modes on ourselves...
	if strings.ToUpper(c.u.Nick()) == strings.ToUpper(string(params[0])) {
		changes, err := UserModes.ParseModeLine(u, u, params[1], params[2:])
		if err != nil {
			c.WriteTo(nil, "501", "%s", err)
		}
		
		prev := &changes
		for cha := changes; cha != nil; cha = cha.Next {
			ok, err := perm.CheckUserData(u, u, cha.Name, cha.Data)
			if !ok {
				m := UserModes.GetMode(cha.Name)
				c.WriteTo(nil, "482", "%s %c: %s", u.Nick(), m,
				          err)
				(*prev) = cha.Next
			} else {
				prev = &cha.Next
			}
		}
		if changes != nil {
			c.u.SetDataList(u, changes)
		}
		return
	}

	if params[0][0] == '#' {
		channame := string(params[0][1:])
		ch := core.FindChannel("", channame)
		if ch == nil {
			return
		}
		changes, err := ChanModes.ParseModeLine(u, ch, params[1], params[2:])
		if err != nil {
			c.WriteTo(nil, "501", "%s", err)
		}
		prev := &changes
		for cha := changes; cha != nil; cha = cha.Next {
			if cha.Member != nil {
				ok, err := perm.CheckMemberData(u, cha.Member, cha.Name, cha.Data)
				if !ok {
					m := ChanModes.GetMode(cha.Name)
					c.WriteTo(nil, "482", "#%s %c: %s", ch.Name(), m, err)
					(*prev) = cha.Next
				} else {
					prev = &cha.Next
				}
			} else {
				ok, err := perm.CheckChanData(u, ch, cha.Name, cha.Data)
				if !ok {
					m := ChanModes.GetMode(cha.Name)
					c.WriteTo(nil, "482", "#%s %c: %s", ch.Name(), m, err)
					(*prev) = cha.Next
				} else {
					prev = &cha.Next
				}
			}
		}
		if changes != nil {
			ch.SetDataList(u, changes)
		}

		return
	}

	c.WriteTo(nil, "501", "%s %s :%s", u.Nick(), params[0],
		  "No such nick or channel.")
}

func cmdPrivmsg(u *core.User, w io.Writer, params [][]byte) {
	c := w.(*Client)

	targets := strings.Split(string(params[0]), ",", -1)
	for _, t := range targets {
		if target := core.GetUserByNick(string(t)); target != nil {
			if ok, err := perm.CheckUserMsg(u, target, params[1], ""); ok {
				target.Message(u, params[1], "")
			} else {
				c.WriteTo(nil, "404", "%s :%s", target.Nick(), err)
			}
			continue
		}

		if t[0] == '#' {
			channame := string(t[1:])
			ch := core.FindChannel("", channame)
			if ch != nil {
				if ok, err := perm.CheckChanMsg(u, ch,
				                                params[1], ""); ok {
					ch.Message(u, params[1], "")
				} else {
					c.WriteTo(nil, "404", "#%s :%s", ch.Name(), err)
				}
				continue
			}
		}

		c.WriteTo(nil, "404", "%s :%s", t, "No such nick or channel.")
	}
}

func cmdNotice(u *core.User, w io.Writer, params [][]byte) {
	c := w.(*Client)

	targets := strings.Split(string(params[0]), ",", -1)
	for _, t := range targets {
		if target := core.GetUserByNick(string(t)); target != nil {
			if ok, err := perm.CheckUserMsg(u, target, params[1],
			                                "noreply"); ok {
				target.Message(u, params[1], "noreply")
			} else {
				c.WriteTo(nil, "404", "%s :%s", target.Nick(), err)
			}
			continue
		}

		if t[0] == '#' {
			channame := string(t[1:])
			ch := core.FindChannel("", channame)
			if ch != nil {
				if ok, err := perm.CheckChanMsg(u, ch,
						params[1]," noreply"); ok {
					ch.Message(u, params[1], "noreply")
				} else {
					c.WriteTo(nil, "404", "#%s :%s", ch.Name(), err)
				}
				continue
			}
		}

		c.WriteTo(nil, "404", "%s :%s", t, "No such nick or channel.")
	}
}

func cmdTopic(u *core.User, w io.Writer, params [][]byte) {
	c := w.(*Client)

	var ch *core.Channel
	if params[0][0] == '#' {
		channame := string(params[0][1:])
		ch = core.FindChannel("", channame)
	}
	if ch == nil {
		c.WriteTo(nil, "404", "%s %s :No such channel.", u.Nick(),
		          params[0])
		return
	}

	// If we're displaying the topic...
	if len(params) < 2 {
		topic, setby, setat := ch.GetTopic()
		if topic != "" {
			c.WriteTo(nil, "332", "#%s :%s", ch.Name(), topic)
			c.WriteTo(nil, "333", "#%s %s %s", ch.Name(), setby,
			          setat)
		} else {
			c.WriteTo(nil, "331", "#%s :No topic is set.",
				  ch.Name())
		}
		return
	}

	// Otherwise, we're setting the topic.
	ch.SetTopic(u, string(params[1]))
}
