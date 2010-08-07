package client

import "fmt"
import "io"

import "oddircd/src/core"
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

	if err := u.SetNick(nick); err != nil {
		fmt.Fprintf(w, ":Server.name 433 %s %s :%s\r\n", u.Nick(),
		            nick, err)
	}
}

func cmdUser(u *core.User, w io.Writer, params [][]byte) {
	if (u.Data("ident") != "") { return }
	
	u.SetData("ident", string(params[0]))
	u.SetData("realname", string(params[3]))
}
