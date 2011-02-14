package client

import "strings"

import "oddcomm/src/core"
import "oddcomm/lib/irc"
import "oddcomm/lib/perm"


// Add core oper commands.
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

func cmdKill(source interface{}, params [][]byte) {
	c := source.(*Client)

	targets := strings.Split(string(params[0]), ",", -1)
	message := string(params[1])
	for _, t := range targets {

		if target := core.GetUserByNick(string(t)); target != nil {
			perm, err := perm.CheckKillPerm(me, c.u, target)
			if perm < -1000000 {
				c.SendLineTo(nil, "404", "%s :%s", target.Nick(), err)
				continue
			}

			// Send a kill message to the user if they're killing
			// themselves. The rest of the server treats this as a
			// quit, but their client won't understand this.
			if target == c.u {
				c.SendLineTo(c.u, "KILL", "%s (%s)", c.u.Nick(), message)
			}

			// Kill the user.
			target.Delete(me, c.u, message)

			continue
		}

		c.SendLineTo(nil, "401", "%s :%s", t, "No such nick.")
	}
}

func cmdDie(source interface{}, params [][]byte) {
	core.Shutdown()
}
