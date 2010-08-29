package client

import "io"

import "oddcomm/src/core"
import "oddcomm/lib/irc"
import "oddcomm/lib/perm"


func init() {
	var c *irc.Command
	if Commands == nil {
		Commands = irc.NewCommandDispatcher()
	}

	c = new(irc.Command)
	c.Name = "WHO"; c.Handler = cmdWho
	c.Minargs = 1; c.Maxargs = 1
	Commands.Add(c)

	c = new(irc.Command)
	c.Name = "NAMES"; c.Handler = cmdNames
	c.Minargs = 1; c.Maxargs = 1
	Commands.Add(c)
	
	c = new(irc.Command)
	c.Name = "OPFLAGS"; c.Handler = cmdOpflags
	c.Minargs = 2; c.Maxargs = 2
	Commands.Add(c)
	
	c = new(irc.Command)
	c.Name = "OPERFLAGS"; c.Handler = cmdOperflags
	c.Minargs = 1; c.Maxargs = 1
	c.OperFlag = "viewflags"
	Commands.Add(c)
}

func cmdWho(u *core.User, w io.Writer, params [][]byte) {
	c := w.(*Client)
	channame := string(params[0])
	if channame[0] == '#' {
		channame = channame[1:]
	}

	var ch *core.Channel
	if ch = core.FindChannel("", channame); ch == nil {
		c.WriteTo(nil, "404", "#%s :No such channel.", channame)
		return
	}	

	// If the user isn't on the channel, don't let them check unless they
	// can view private channel data.
	if m := ch.GetMember(u); m == nil {
		if ok, err := perm.CheckChanViewData(u, ch, "members"); !ok {
			c.WriteTo(nil, "482", "#%s :%s", ch.Name(), err)
			return
		}
	}

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

func cmdNames(u *core.User, w io.Writer, params [][]byte) {
	c := w.(*Client)
	channame := string(params[0])
	if channame[0] == '#' {
		channame = channame[1:]
	}

	var ch *core.Channel
	if ch = core.FindChannel("", channame); ch == nil {
		c.WriteTo(nil, "404", "#%s :No such channel.", channame)
		return
	}

	// If the user isn't on the channel, don't let them check unless they
	// can view private channel data.
	if m := ch.GetMember(u); m == nil {
		if ok, err := perm.CheckChanViewData(u, ch, "members"); !ok {
			c.WriteTo(nil, "482", "#%s :%s", ch.Name(), err)
			return
		}
	}

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

func cmdOpflags(u *core.User, w io.Writer, params [][]byte) {
	c := w.(*Client)
	channame := string(params[0])
	if channame[0] == '#' {
		channame = channame[1:]
	}

	var ch *core.Channel
	if ch = core.FindChannel("", channame); ch == nil {
		c.WriteTo(nil, "404", "#%s :No such channel.", channame)
		return
	}

	// If the user isn't on the channel, don't let them check unless they
	// can view private channel data.
	if m := ch.GetMember(u); m == nil {
		if ok, err := perm.CheckChanViewData(u, ch, "members"); !ok {
			c.WriteTo(nil, "482", "#%s :%s", ch.Name(), err)
			return
		}
	}

	var target *core.User
	if target = core.GetUserByNick(string(params[1])); target == nil {
		c.WriteTo(nil, "404", "%s :No such user.", params[1])
		return
	}

	var m *core.Membership
	if m = ch.GetMember(target); m == nil {
		c.WriteTo(nil, "304", ":OPFLAGS #%s: %s is not in the channel.", ch.Name(), target.Nick())
		return
	}
	if ok, err := perm.CheckMemberViewData(u, m, "op"); !ok {
		c.WriteTo(nil, "482", "#%s :%s: %s", ch.Name(), target.Nick(), err)
		return
	}

	var flags string
	if flags = m.Data("op"); flags == "" {
		c.WriteTo(nil, "304", ":OPFLAGS #%s: %s has no channel op flags.", ch.Name(), target.Nick())
		return
	}

	if flags == "on" {
		flags = perm.DefaultChanOp()
	}	
	c.WriteTo(nil, "304", ":OPFLAGS #%s: %s has channel op flags: %s", ch.Name(), target.Nick(), flags)
}

func cmdOperflags(u *core.User, w io.Writer, params [][]byte) {
	c := w.(*Client)

	var target *core.User
	if target = core.GetUserByNick(string(params[0])); target == nil {
		c.WriteTo(nil, "404", "%s %s :No such user.", u.Nick(),
		          params[0])
		return
	}

	var flags string
	if flags = target.Data("op"); flags == "" {
		c.WriteTo(nil, "304", ":OPERFLAGS: %s has no server oper flags.", target.Nick())
		return
	}

	if flags == "on" {
		flags = perm.DefaultServerOp()
	}	
	c.WriteTo(nil, "304", ":OPERFLAGS: %s has server oper flags: %s", target.Nick(), flags)

	var commands string
	if commands = target.Data("opcommands"); commands == "" {
		return
	}
	c.WriteTo(nil, "304", ":OPERFLAGS: %s also has the following specific commands: %s", target.Nick(), commands)
}
