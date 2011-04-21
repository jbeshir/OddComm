/*
	Provides a command for password-based authentication to an account.
*/
package login

import "strings"

import "oddcomm/src/client"
import "oddcomm/lib/irc"
import "oddcomm/lib/perm"


func init() {
	var c *irc.Command

	// Add login command.
	c = new(irc.Command)
	c.Name = "LOGIN"
	c.Handler = cmdLogin
	c.Minargs = 1
	c.Maxargs = 2
	c.Unregged = 1
	client.Commands.Add(c)

	// Add pass command, just an alias.
	c = new(irc.Command)
	c.Name = "PASS"
	c.Handler = cmdLogin
	c.Minargs = 1
	c.Maxargs = 2
	c.Unregged = 1
	client.Commands.Add(c)

	// Add identify command, just an alias.
	c = new(irc.Command)
	c.Name = "IDENTIFY"
	c.Handler = cmdLogin
	c.Minargs = 1
	c.Maxargs = 2
	c.Unregged = 1
	client.Commands.Add(c)
}

func cmdLogin(source interface{}, params [][]byte) {
	c := source.(*client.Client)

	var account string
	var pass string

	// If we've only got one parameter, if it has a colon, split it there
	// and take the prefix as an account name, and if it lacks one, take
	// their nick as the account name.
	if len(params) == 1 {
		pass = string(params[0])
		colon := strings.IndexRune(pass, ':')
		if colon != -1 && colon < len(pass)-1 {
			account = pass[:colon]
			pass = pass[colon+1:]
		} else {
			account = c.User().Nick()
		}
	} else {
		account = string(params[0])
		pass = string(params[1])
	}

	// Try to log them in.
	if ok, err := perm.CheckLogin("", c.User(), account, "password", pass); ok {
		c.User().SetData(nil, nil, "account", err.String())
	} else {
		c.SendLineTo(nil, "491", ":%s", err)
	}
}
