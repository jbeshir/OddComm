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
		c.WriteTo(nil, "304", ":%s has no server operator flags.", target.Nick())
		return
	}

	if flags == "on" {
		flags = perm.DefaultServerOp()
	}	
	c.WriteTo(nil, "304", ":%s has server operator flags: %s", target.Nick(), flags)

	var commands string
	if commands = target.Data("opcommands"); commands == "" {
		return
	}
	c.WriteTo(nil, "304", ":%s also has the following specific commands: %s", target.Nick(), commands)
}
