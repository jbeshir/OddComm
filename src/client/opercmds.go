package client

import "io"

import "oddcomm/src/core"
import "oddcomm/lib/irc"


// Add core oper commands.
func init() {
	var c *irc.Command
	if Commands == nil {
		Commands = irc.NewCommandDispatcher()
	}

	c = new(irc.Command)
	c.Name = "DIE"; c.Handler = cmdDie
	c.OperFlag = "shutdown"
	Commands.Add(c)
}

func cmdDie(u *core.User, w io.Writer, params [][]byte) {
	core.Shutdown()
}
