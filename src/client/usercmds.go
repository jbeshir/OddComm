package client

import "io"

import "oddircd/src/core"
import "oddircd/src/perm"
import "oddircd/src/irc"


// Add core user commands.
func init() {
	var c *irc.Command

	c = new(irc.Command)
	c.Handler = cmdNick
	c.Minargs = 1
	c.Maxargs = 1
	c.Unregged = 1
	Commands.Add("NICK", c)

	c = new(irc.Command)
	c.Handler = cmdUser
	c.Minargs = 4
	c.Maxargs = 4
	c.Unregged = 2
	Commands.Add("USER", c)

	c = new(irc.Command)
	c.Handler = cmdPrivmsg
	c.Minargs = 2
	c.Maxargs = 2
	c.Unregged = 0
	Commands.Add("PRIVMSG", c)

	c = new(irc.Command)
	c.Handler = cmdNotice
	c.Minargs = 2
	c.Maxargs = 2
	c.Unregged = 0
	Commands.Add("NOTICE", c)

	c = new(irc.Command)
	c.Handler = irc.CmdQuit
	c.Maxargs = 1
	c.Unregged = 1
	Commands.Add("QUIT", c)
}

func cmdNick(u *core.User, w io.Writer, params [][]byte) {
	var nick = string(params[0])

	if nick == "0" {
		nick = u.ID()		
	}

	if !perm.ValidateNick(u, nick) {
		return
	}

	if err := u.SetNick(nick); err != nil {
		if c, ok := w.(*Client); ok {
			c.WriteFrom(nil, "433", "%s :%s", nick, err)
		}
	}
}

func cmdUser(u *core.User, w io.Writer, params [][]byte) {
	if (u.Data("ident") != "") { return }

	ident := string(params[0])
	realname := string(params[3])

	if !perm.ValidateIdent(u, ident) {
		return
	}

	if !perm.ValidateRealname(u, realname) {
		return
	}
	
	u.SetData("ident", string(params[0]))
	u.SetData("realname", string(params[3]))
}

func cmdNotice(u *core.User, w io.Writer, params [][]byte) {
	target := core.GetUserByNick(string(params[0]))
	if target != nil {
		if perm.CheckPM(u, target, params[1], "reply") {
			target.PM(u, params[1], "reply")
		}
	}
}

func cmdPrivmsg(u *core.User, w io.Writer, params [][]byte) {
	target := core.GetUserByNick(string(params[0]))
	if target != nil {
		if perm.CheckPM(u, target, params[1], "") {
			target.PM(u, params[1], "")
		}
	}
}
