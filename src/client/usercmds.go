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
	c.Handler = cmdMode
	c.Minargs = 2
	c.Maxargs = 42
	Commands.Add("MODE", c)

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
			c.WriteFrom(nil, "432", "%s :%s", nick, err)
		}
		return
	}

	if err := u.SetNick(nick); err != nil {
		if c, ok := w.(*Client); ok {
			c.WriteFrom(nil, "433", "%s :%s", nick, err)
		}
	}
}

func cmdUser(u *core.User, w io.Writer, params [][]byte) {
	c := w.(*Client)
	if (u.Data("ident") != "") { return }

	ident := string(params[0])
	realname := string(params[3])

	if ok, err := perm.ValidateIdent(u, ident); !ok {
		c.WriteFrom(nil, "461", "USER :%s", err)
		return
	}

	if ok, err := perm.ValidateRealname(u, realname); !ok {
		c.WriteFrom(nil, "461", "USER :%s", err)
		return
	}
	
	u.SetData("ident", string(params[0]))
	u.SetData("realname", string(params[3]))
}

func cmdPing(u *core.User, w io.Writer, params [][]byte) {
	c := w.(*Client)
	c.WriteServer("PONG %s :%s", "Server.name", params[0])
	
}

func cmdMode(u *core.User, w io.Writer, params [][]byte) {
	c := w.(*Client)
	if strings.ToUpper(c.u.Nick()) == strings.ToUpper(string(params[0])) {
		changes, err := UserModes.ParseModeLine(u, params[1], params[2:])
		if err != nil {
			c.WriteFrom(nil, "501", "%s", err)
		}
		if changes != nil {
			c.u.SetDataList(changes)
		}
	} else {
		c.WriteFrom(nil, "501", "%s %s :%s", u.Nick(), params[0],
		            "No such nick/channel")
	}
}

func cmdPrivmsg(u *core.User, w io.Writer, params [][]byte) {
	c := w.(*Client)
	if target := core.GetUserByNick(string(params[0])); target != nil {
		if ok, err := perm.CheckPM(u, target, params[1], ""); ok {
			target.PM(u, params[1], "")
		} else {
			c.WriteFrom(nil, "404", "%s %s :%s", u.Nick(),
			            target.Nick(), err)
		}
		return
	}
}

func cmdNotice(u *core.User, w io.Writer, params [][]byte) {
	c := w.(*Client)
	if target := core.GetUserByNick(string(params[0])); target != nil {
		if ok, err := perm.CheckPM(u, target, params[1], "reply"); ok {
			target.PM(u, params[1], "reply")
		} else {
			c.WriteFrom(nil, "404", "%s %s :%s", u.Nick(),
			            target.Nick(), err)
		}
		return
	}
}
