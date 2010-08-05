package client

import "oddircd/irc"


// Add core user commands.
func init() {
	var c *irc.Command

	c = new(irc.Command)
	c.Handler = irc.Nick
	c.Minargs = 1
	c.Maxargs = 1
	c.Unregged = true
	Commands.Add("NICK", c)

	c = new(irc.Command)
	c.Handler = irc.Quit
	c.Maxargs = 1
	c.Unregged = true
	Commands.Add("QUIT", c)
}
