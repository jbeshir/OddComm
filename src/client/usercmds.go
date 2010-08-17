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
	c.Handler = cmdNames
	c.Minargs = 1
	c.Maxargs = 1
	Commands.Add("NAMES", c)

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

	if ok, err := perm.ValidateNick(u, nick); !ok {
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

	ident := string(params[0])
	realname := string(params[3])

	if ok, err := perm.ValidateIdent(u, ident); !ok {
		c.WriteTo(nil, "461", "USER :%s", err)
		return
	}

	if ok, err := perm.ValidateRealname(u, realname); !ok {
		c.WriteTo(nil, "461", "USER :%s", err)
		return
	}
	
	u.SetData(nil, "ident", "~" + string(params[0]))
	u.SetData(nil, "realname", string(params[3]))
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
			c.WriteTo(nil, "352", "#%s %s %s %s %s H* :0 %s",
			          channame, user.GetIdent(),
			          user.GetHostname(), "Server.name",
			          user.Nick(), user.Data("realname"))
		}
		c.WriteTo(nil, "315", "#%s :End of /WHO list.", channame)
	}
}


func cmdJoin(u *core.User, w io.Writer, params [][]byte) {
	channame := string(params[0])
	if channame[0] == '#' {
		channame = channame[1:]
	}
	
	core.GetChannel("", channame).Join(u)

	cmdNames(u, w, params[0:1])
}

func cmdPart(u *core.User, w io.Writer, params [][]byte) {
	channame := string(params[0])
	if channame[0] == '#' {
		channame = channame[1:]
	}

	if ch := core.FindChannel("", channame); ch != nil {
		ch.Remove(u, u)
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
			if names != "" {
				names += " "
			}
			names += it.User().Nick()
		}
		c.WriteTo(nil, "353", "= #%s :%s", channame, names)
		c.WriteTo(nil, "366", "#%s :End of /NAMES list", channame)
	}
}

func cmdMode(u *core.User, w io.Writer, params [][]byte) {
	c := w.(*Client)

	// If we're viewing the modes of a user or channel...
	if len(params) < 2 {
		if strings.ToUpper(c.u.Nick()) == strings.ToUpper(string(params[0])) {
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
			// Different, fixed numerics for different
			// modes. Stupid protocol.
			num := "941"; endnum := "940"
			switch mode {
			case 'b': num = "367"; endnum = "368"
			case 'e': num = "348"; endnum = "349"
			case 'I': num = "346"; endnum = "347"
			}

			valid := ChanModes.ListMode(ch, int(mode),
			                   func(p, v string) {
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
			if valid {
				c.WriteTo(nil, endnum,
				          "#%s :End of mode list.",
				          ch.Name())
			} else {
				badmodes += string(mode)
			}
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


	if strings.ToUpper(c.u.Nick()) == strings.ToUpper(string(params[0])) {
		changes, err := UserModes.ParseModeLine(u, u, params[1], params[2:])
		if err != nil {
			c.WriteTo(nil, "501", "%s", err)
		}
		if changes != nil {
			c.u.SetDataList(u, changes)
		}
		return
	}

	if params[0][0] == '#' {
		channame := string(params[0][1:])
		ch := core.FindChannel("", channame)
		if ch != nil {
			changes, err := ChanModes.ParseModeLine(u, ch, params[1], params[2:])
			if err != nil {
				c.WriteTo(nil, "501", "%s", err)
			}
			if changes != nil {
				ch.SetDataList(u, changes)
			}
		}
		return
	}

	c.WriteTo(nil, "501", "%s %s :%s", u.Nick(), params[0],
		  "No such nick/channel")
}

func cmdPrivmsg(u *core.User, w io.Writer, params [][]byte) {
	c := w.(*Client)
	if target := core.GetUserByNick(string(params[0])); target != nil {
		if ok, err := perm.CheckPM(u, target, params[1], ""); ok {
			target.Message(u, params[1], "")
		} else {
			c.WriteTo(nil, "404", "%s %s :%s", u.Nick(),
			            target.Nick(), err)
		}
		return
	}

	if params[0][0] == '#' {
		channame := string(params[0][1:])
		ch := core.FindChannel("", channame)
		if ch != nil {
			ch.Message(u, params[1], "")
		}
		return
	}
}

func cmdNotice(u *core.User, w io.Writer, params [][]byte) {
	c := w.(*Client)
	if target := core.GetUserByNick(string(params[0])); target != nil {
		if ok, err := perm.CheckPM(u, target, params[1],
				"noreply"); ok {
			target.Message(u, params[1], "noreply")
		} else {
			c.WriteTo(nil, "404", "%s %s :%s", u.Nick(),
			          target.Nick(), err)
		}
		return
	}

	if params[0][0] == '#' {
		channame := string(params[0][1:])
		ch := core.FindChannel("", channame)
		if ch != nil {
			ch.Message(u, params[1], "noreply")
		}
		return
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
