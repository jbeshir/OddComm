package client

import "fmt"
import "strings"
import "time"

import "oddcomm/src/core"
import "oddcomm/lib/irc"
import "oddcomm/lib/perm"


func init() {
	var c *irc.Command
	if Commands == nil {
		Commands = irc.NewCommandDispatcher()
	}

	c = new(irc.Command)
	c.Name = "VERSION"
	c.Handler = cmdVersion
	Commands.Add(c)

	c = new(irc.Command)
	c.Name = "USERHOST"
	c.Handler = cmdUserhost
	c.Minargs = 1
	c.Maxargs = 1
	Commands.Add(c)

	c = new(irc.Command)
	c.Name = "ISON"
	c.Handler = cmdIsOn
	c.Minargs = 1
	c.Maxargs = 1
	Commands.Add(c)

	c = new(irc.Command)
	c.Name = "WHO"
	c.Handler = cmdWho
	c.Minargs = 1
	c.Maxargs = 1
	Commands.Add(c)

	c = new(irc.Command)
	c.Name = "NAMES"
	c.Handler = cmdNames
	c.Minargs = 1
	c.Maxargs = 1
	Commands.Add(c)

	c = new(irc.Command)
	c.Name = "TIME"
	c.Handler = cmdTime
	Commands.Add(c)

	c = new(irc.Command)
	c.Name = "OPFLAGS"
	c.Handler = cmdOpflags
	c.Minargs = 2
	c.Maxargs = 2
	Commands.Add(c)

	c = new(irc.Command)
	c.Name = "OPERFLAGS"
	c.Handler = cmdOperflags
	c.Minargs = 1
	c.Maxargs = 1
	c.OperFlag = "viewflags"
	Commands.Add(c)
}


func cmdVersion(source interface{}, params [][]byte) {
	c := source.(*Client)

	c.SendLineTo(nil, "351", "OddComm-%s Server.name", core.Version)
	c.SendLineTo(nil, "351", "%s :are supported by this server", supportLine)
}

func cmdUserhost(source interface{}, params [][]byte) {
	c := source.(*Client)

	nicks := strings.Fields(string(params[0]))
	var replyline string
	for _, nick := range nicks {
		var user *core.User
		if user = core.GetUserByNick(nick); user == nil {
			continue
		}

		if replyline != "" {
			replyline += " "
		}

		replyline += nick
		if user.Data("op") != "" {
			replyline += "*"
		}
		replyline += "="
		if user.Data("away") != "" {
			replyline += "-"
		} else {
			replyline += "+"
		}
		replyline += user.GetHostname()
	}

	c.SendLineTo(nil, "302", ":%s", replyline)
}

func cmdIsOn(source interface{}, params [][]byte) {
	c := source.(*Client)

	nicks := strings.Fields(string(params[0]))
	var replyline string
	for _, nick := range nicks {
		if core.GetUserByNick(nick) != nil {
			if replyline != "" {
				replyline += " "
			}
			replyline += nick
		}
	}

	c.SendLineTo(nil, "303", ":%s", replyline)
}

func cmdWho(source interface{}, params [][]byte) {
	c := source.(*Client)
	channame := string(params[0])
	if channame[0] == '#' {
		channame = channame[1:]
	}

	var ch *core.Channel
	if ch = core.FindChannel("", channame); ch == nil {
		c.SendLineTo(nil, "403", "#%s :No such channel.", channame)
		return
	}

	// If the user isn't on the channel, don't let them check unless they
	// can view private channel data.
	if m := ch.GetMember(c.u); m == nil {
		if ok, err := perm.CheckChanViewData(me, c.u, ch, "members"); !ok {
			c.SendLineTo(nil, "482", "#%s :%s", ch.Name(), err)
			return
		}
	}

	it := ch.Users()
	c.WriteBlock(func() []byte {
		if it == nil {
			return nil
		}

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

		result := fmt.Sprintf(":%s 352 %s #%s %s %s %s %s %s :0 %s\r\n",
			"Server.name", c.u.Nick(), channame, user.GetIdent(),
			user.GetHostname(), "Server.name", user.Nick(),
			prefixes, user.Data("realname"))

		it = it.ChanNext()
		return []byte(result)
	})

	c.SendLineTo(nil, "315", "%s :End of /WHO list.", params[0])
}

func cmdNames(source interface{}, params [][]byte) {
	c := source.(*Client)
	channame := string(params[0])
	if len(channame) > 0 && channame[0] == '#' {
		channame = channame[1:]
	}

	var ch *core.Channel
	if ch = core.FindChannel("", channame); ch == nil {
		c.SendLineTo(nil, "403", "#%s :No such channel.", channame)
		return
	}

	// If the user isn't on the channel, don't let them check unless they
	// can view private channel data.
	// Otherwise, get their prefixes.
	var myprefix string
	if m := ch.GetMember(c.u); m == nil {
		if ok, err := perm.CheckChanViewData(me, c.u, ch, "members"); !ok {
			c.SendLineTo(nil, "482", "#%s :%s", ch.Name(), err)
			return
		}
		myprefix = "="
	} else {
		myprefix = ChanModes.GetPrefixes(m)
	}

	it := ch.Users()
	c.WriteBlock(func() []byte {
		if it == nil {
			return nil
		}

		names := fmt.Sprintf(":%s 353 %s %s #%s :", "Server.name",
			c.u.Nick(), myprefix, channame)

		for ; it != nil; it = it.ChanNext() {
			name := ChanModes.GetPrefixes(it) + it.User().Nick()
			if len(names)+len(name) > 508 {
				break
			}
			names += " " + name
		}

		names += "\r\n"
		return []byte(names)
	})

	c.SendLineTo(nil, "366", "#%s :End of /NAMES list", channame)
}

func cmdTime(source interface{}, params [][]byte) {
	c := source.(*Client)
	c.SendLineTo(nil, "391", ":%s", time.UTC().Format(time.RFC1123))
}

func cmdOpflags(source interface{}, params [][]byte) {
	c := source.(*Client)
	channame := string(params[0])
	if channame[0] == '#' {
		channame = channame[1:]
	}

	var ch *core.Channel
	if ch = core.FindChannel("", channame); ch == nil {
		c.SendLineTo(nil, "403", "#%s :No such channel.", channame)
		return
	}

	// If the user isn't on the channel, don't let them check unless they
	// can view private channel data.
	if m := ch.GetMember(c.u); m == nil {
		if ok, err := perm.CheckChanViewData(me, c.u, ch, "members"); !ok {
			c.SendLineTo(nil, "482", "#%s :%s", ch.Name(), err)
			return
		}
	}

	var target *core.User
	if target = core.GetUserByNick(string(params[1])); target == nil {
		c.SendLineTo(nil, "401", "%s :No such user.", params[1])
		return
	}

	var m *core.Membership
	if m = ch.GetMember(target); m == nil {
		c.SendLineTo(nil, "304", ":OPFLAGS #%s: %s is not in the channel.", ch.Name(), target.Nick())
		return
	}
	if ok, err := perm.CheckMemberViewData(me, c.u, m, "op"); !ok {
		c.SendLineTo(nil, "482", "#%s :%s: %s", ch.Name(), target.Nick(), err)
		return
	}

	var flags string
	if flags = m.Data("op"); flags == "" {
		c.SendLineTo(nil, "304", ":OPFLAGS #%s: %s has no channel op flags.", ch.Name(), target.Nick())
		return
	}

	if flags == "on" {
		flags = perm.DefaultChanOp()
	}
	c.SendLineTo(nil, "304", ":OPFLAGS #%s: %s has channel op flags: %s", ch.Name(), target.Nick(), flags)
}

func cmdOperflags(source interface{}, params [][]byte) {
	c := source.(*Client)

	var target *core.User
	if target = core.GetUserByNick(string(params[0])); target == nil {
		c.SendLineTo(nil, "401", "%s %s :No such user.", c.u.Nick(),
			params[0])
		return
	}

	var flags string
	if flags = target.Data("op"); flags == "" {
		c.SendLineTo(nil, "304", ":OPERFLAGS: %s has no server oper flags.", target.Nick())
		return
	}

	if flags == "on" {
		flags = perm.DefaultServerOp()
	}
	c.SendLineTo(nil, "304", ":OPERFLAGS: %s has server oper flags: %s", target.Nick(), flags)

	var commands string
	if commands = target.Data("opcommands"); commands == "" {
		return
	}
	c.SendLineTo(nil, "304", ":OPERFLAGS: %s also has the following specific commands: %s", target.Nick(), commands)
}
