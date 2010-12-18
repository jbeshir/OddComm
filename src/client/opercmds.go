package client

import "io"
import "strings"

import "oddcomm/src/core"
import "oddcomm/lib/irc"
import "oddcomm/lib/perm"


// Add cUore oper commands.
func init() {
	var c *irc.Command
	if Commands == nil {
		Commands = irc.NewCommandDispatcher()
	}

	c = new(irc.Command)
	c.Name = "KILL"
	c.Handler = cmdKill
	c.Minargs = 2
	c.Maxargs = 2
	c.OperFlag = "ban"
	Commands.Add(c)

	c = new(irc.Command)
	c.Name = "DIE"
	c.Handler = cmdDie
	c.OperFlag = "shutdown"
	Commands.Add(c)
}

func cmdKill(u *core.User, w io.Writer, params [][]byte) {
	c := w.(*Client)

	targets := strings.Split(string(params[0]), ",", -1)
	message := string(params[1])
	for _, t := range targets {

		if target := core.GetUserByNick(string(t)); target != nil {
			perm, err := perm.CheckKillPerm(u, target)
			if perm < -1000000 {
				c.WriteTo(nil, "404", "%s :%s", target.Nick(), err)
				continue
			}

			// Send a kill message to the user if they're killing
			// themselves. The rest of the server treats this as a
			// quit, but their client won't understand this.
			if target == u {
				c.WriteTo(u, "KILL", "%s (%s)", u.Nick(), message)
			}

			// Kill the user.
			target.Delete(u, message)

			continue
		}

		c.WriteTo(nil, "401", "%s :%s", t, "No such nick.")
	}
}

func cmdDie(u *core.User, w io.Writer, params [][]byte) {
	core.Shutdown()
}
