package client

import "strconv"
import "strings"
import "time"

import "oddcomm/src/core"
import "oddcomm/lib/perm"
import "oddcomm/lib/irc"


// Add core user commands.
func init() {
	var c *irc.Command
	if Commands == nil {
		Commands = irc.NewCommandDispatcher()
	}

	c = new(irc.Command)
	c.Name = "USER"
	c.Handler = cmdUser
	c.Minargs = 4
	c.Maxargs = 4
	c.Unregged = 2
	Commands.Add(c)

	c = new(irc.Command)
	c.Name = "NICK"
	c.Handler = cmdNick
	c.Minargs = 1
	c.Maxargs = 1
	c.Unregged = 1
	Commands.Add(c)

	c = new(irc.Command)
	c.Name = "AWAY"
	c.Handler = cmdAway
	c.Maxargs = 1
	Commands.Add(c)

	c = new(irc.Command)
	c.Name = "QUIT"
	c.Handler = cmdQuit
	c.Maxargs = 1
	c.Unregged = 1
	Commands.Add(c)

	c = new(irc.Command)
	c.Name = "PING"
	c.Handler = cmdPing
	c.Minargs = 1
	c.Maxargs = 1
	Commands.Add(c)

	c = new(irc.Command)
	c.Name = "MODE"
	c.Handler = cmdMode
	c.Minargs = 1
	c.Maxargs = 3
	Commands.Add(c)

	c = new(irc.Command)
	c.Name = "PRIVMSG"
	c.Handler = cmdPrivmsg
	c.Minargs = 2
	c.Maxargs = 2
	Commands.Add(c)

	c = new(irc.Command)
	c.Name = "NOTICE"
	c.Handler = cmdNotice
	c.Minargs = 2
	c.Maxargs = 2
	Commands.Add(c)
}

func cmdNick(source interface{}, params [][]byte) {
	c := source.(*Client)

	var nick = string(params[0])

	if nick == "0" {
		nick = c.u.ID()
	}

	if ok, err := perm.CheckNick(me, c.u, nick); !ok {
		if c, ok := source.(*Client); ok {
			c.WriteTo(nil, "432", "%s :%s", nick, err)
		}
		return
	}

	if err := c.u.SetNick(me, nick, -1); err != nil {
		if c, ok := source.(*Client); ok {
			c.WriteTo(nil, "433", "%s :%s", nick, err)
		}
	}

	// Track whether we've gotten a NICK command from this client yet.
	if c, ok := source.(*Client); ok && !c.nicked {
		c.nicked = true
		c.u.PermitRegistration(me)
	}
}

func cmdUser(source interface{}, params [][]byte) {
	c := source.(*Client)
	if c.u.Data("ident") != "" {
		return
	}

	ident := "~" + string(params[0])
	real := string(params[3])

	// Check that the ident and realname are valid.
	if num, err := perm.CheckUserDataPerm(me, c.u, c.u, "ident", ident); num <= -1e9 {
		c.WriteTo(nil, "461", "USER :%s", err)
		return
	}
	if num, err := perm.CheckUserDataPerm(me, c.u, c.u, "realname", real); num <= -1e9 {
		c.WriteTo(nil, "461", "USER :%s", err)
		return
	}

	data := make([]core.DataChange, 2)
	data[0].Name, data[0].Data = "ident", ident
	data[1].Name, data[1].Data = "real", real
	c.u.SetDataList(me, nil, data)
}

func cmdPing(source interface{}, params [][]byte) {
	c := source.(*Client)
	c.WriteFrom(nil, "PONG %s :%s", "Server.name", params[0])

}

func cmdMode(source interface{}, params [][]byte) {
	c := source.(*Client)

	// If we're viewing the modes of ourselves or a channel...
	if len(params) < 2 {
		if strings.ToUpper(c.u.Nick()) == strings.ToUpper(string(params[0])) {
			modeline := UserModes.GetModes(c.u)
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

			if ok, err := perm.CheckChanViewData(me, c.u, ch, name); !ok {
				c.WriteTo(nil, "482", "#%s :%s", ch.Name(), err)
				continue
			}

			// Different, fixed numerics for different
			// modes. Stupid protocol.
			num := "941"
			endnum := "940"
			switch mode {
			case 'b':
				num = "367"
				endnum = "368"
			case 'e':
				num = "348"
				endnum = "349"
			case 'I':
				num = "346"
				endnum = "347"
			}

			valid := ChanModes.ListMode(ch, int(mode), func(p, v string) {
				var setTime string = "0"
				var setBy string = "Server.name"
				words := strings.Fields(v)
				for _, word := range words {
					if len(word) > 6 && word[:6] == "setat-" {
						setTime = word[6:]
						continue
					}
					if len(word) > 6 && word[:6] == "setby-" {
						setBy = word[6:]
						continue
					}
				}

				c.WriteTo(nil, num, "#%s %s %s %s",
					ch.Name(), p, setBy, setTime)
			})
			if valid {
				c.WriteTo(nil, endnum, "#%s :End of mode list.",
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

	// If we're setting modes on ourselves...
	if strings.ToUpper(c.u.Nick()) == strings.ToUpper(string(params[0])) {
		var mpars []string
		if len(params) == 3 {
			mpars = strings.Fields(string(params[2]))
		}
		changes, err := UserModes.ParseModeLine(c.u, c.u, params[1], mpars)
		if err != nil {
			c.WriteTo(nil, "501", "%s", err)
		}

		todo := make([]core.DataChange, 0, len(changes))
		for _, cha := range changes {
			ok, err := perm.CheckUserData(me, c.u, c.u, cha.Name, cha.Data)
			if !ok {
				m := UserModes.GetMode(cha.Name)
				c.WriteTo(nil, "482", "%s %c: %s", c.u.Nick(),
					m, err)
				continue
			}
			todo = append(todo, cha)
		}
		if len(todo) != 0 {
			c.u.SetDataList(me, c.u, todo)
		}
		return
	}

	if params[0][0] == '#' {
		channame := string(params[0][1:])
		ch := core.FindChannel("", channame)
		if ch == nil {
			return
		}

		var mpars []string
		if len(params) == 3 {
			mpars = strings.Fields(string(params[2]))
		}
		changes, err := ChanModes.ParseModeLine(c.u, ch, params[1], mpars)
		if err != nil {
			c.WriteTo(nil, "501", "%s", err)
		}

		todo := make([]core.DataChange, 0, len(changes))
		for _, cha := range changes {
			if cha.Member != nil {
				ok, err := perm.CheckMemberData(me, c.u, cha.Member, cha.Name, cha.Data)
				if !ok {
					m := ChanModes.GetMode(cha.Name)
					c.WriteTo(nil, "482", "#%s %c: %s", ch.Name(), m, err)
					continue
				}
			} else {
				ok, err := perm.CheckChanData(me, c.u, ch, cha.Name, cha.Data)
				if !ok {
					m := ChanModes.GetMode(cha.Name)
					c.WriteTo(nil, "482", "#%s %c: %s", ch.Name(), m, err)
					continue
				}
			}
			todo = append(todo, cha)
		}
		if todo != nil {
			ch.SetDataList(me, c.u, todo)
		}

		return
	}

	c.WriteTo(nil, "401", "%s %s :%s", c.u.Nick(), params[0],
		"No such nick or channel.")
}

func cmdPrivmsg(source interface{}, params [][]byte) {
	c := source.(*Client)

	targets := strings.Split(string(params[0]), ",", -1)
	for _, t := range targets {
		if target := core.GetUserByNick(string(t)); target != nil {
			if ok, err := perm.CheckUserMsg(me, c.u, target, params[1], ""); ok {
				if v := target.Data("away"); v != "" {
					c.WriteTo(nil, "301", "%s :%s",
						target.Nick(), v)
				}
				target.Message(me, c.u, params[1], "")
			} else {
				c.WriteTo(nil, "404", "%s :%s", target.Nick(), err)
			}
			continue
		}

		if t[0] == '#' {
			channame := string(t[1:])
			ch := core.FindChannel("", channame)
			if ch != nil {
				if ok, err := perm.CheckChanMsg(c.u, ch,
					params[1], ""); ok {
					ch.Message(me, c.u, params[1], "")
				} else {
					c.WriteTo(nil, "404", "#%s :%s", ch.Name(), err)
				}
				continue
			}
		}

		c.WriteTo(nil, "401", "%s :%s", t, "No such nick or channel.")
	}
}

func cmdNotice(source interface{}, params [][]byte) {
	c := source.(*Client)

	targets := strings.Split(string(params[0]), ",", -1)
	for _, t := range targets {
		if target := core.GetUserByNick(string(t)); target != nil {
			if ok, err := perm.CheckUserMsg(me, c.u, target, params[1],
				"noreply"); ok {
				target.Message(me, c.u, params[1], "noreply")
			} else {
				c.WriteTo(nil, "404", "%s :%s", target.Nick(), err)
			}
			continue
		}

		if t[0] == '#' {
			channame := string(t[1:])
			ch := core.FindChannel("", channame)
			if ch != nil {
				if ok, err := perm.CheckChanMsg(c.u, ch,
					params[1], " noreply"); ok {
					ch.Message(me, c.u, params[1], "noreply")
				} else {
					c.WriteTo(nil, "404", "#%s :%s", ch.Name(), err)
				}
				continue
			}
		}

		c.WriteTo(nil, "401", "%s :%s", t, "No such nick or channel.")
	}
}

func cmdAway(source interface{}, params [][]byte) {
	c := source.(*Client)

	var message string
	if len(params) != 0 {
		message = string(params[0])
	}

	if message == "" {
		c.u.SetData(me, c.u, "away", "")
		c.u.SetData(me, c.u, "away time", "")
		c.WriteTo(nil, "305", ":You are no longer marked as being away.")
	} else {
		c.u.SetData(me, c.u, "away", message)
		c.u.SetData(me, c.u, "away time", strconv.Itoa64(time.Seconds()))
		c.WriteTo(nil, "306", ":You have been marked as being away.")
	}
}


func cmdQuit(source interface{}, params [][]byte) {
	c := source.(*Client)

	if len(params) > 0 {
		c.u.Delete(me, c.u, string(params[0]))
	} else {
		c.u.Delete(me, c.u, "")
	}
}
