package ts6

import "oddcomm/lib/irc"
import "oddcomm/src/core"


var commands irc.CommandDispatcher = irc.NewCommandDispatcher()


func init() {
	var c *irc.Command

	// Add pass command.
	c = new(irc.Command)
	c.Name = "PASS"
	c.Handler = cmdPass
	c.Minargs = 4
	c.Maxargs = 4
	c.Unregged = 1
	commands.Add(c)

	// Add pass command.
	c = new(irc.Command)
	c.Name = "SERVER"
	c.Handler = cmdServer
	c.Minargs = 3
	c.Maxargs = 3
	c.Unregged = 1
	commands.Add(c)
}


// Password authentication.
// May only be from a server.
func cmdPass(source interface{}, params [][]byte) {
	s, ok := source.(*server)
	if !ok {
		return
	}

	// If this isn't a local server, ignore it.
	l := s.local
	if &l.s != s {
		return
	}

	// Ignore PASS commands from already authed servers.
	// They can happen if they have an invalid source.
	if l.authed {
		return
	}

	// Set the password.
	l.pass = string(params[0])

	if l.s.u != nil {
		// Already got a user.
		return
	}

	// Validate the given SID.
	if len(params[3]) != 3 {
		l.c.Close()
		return
	}
	if params[3][0] < '0' || params[3][0] > '9' {
		l.c.Close()
		return
	}
	if (params[3][0] < '0' || params[3][0] > '9') && (params[3][1] < 'A' || params[3][1] > 'Z') {
		l.c.Close()
		return
	}
	if (params[3][0] < '0' || params[3][0] > '9') && (params[3][2] < 'A' || params[3][2] > 'Z') {
		l.c.Close()
		return
	}

	// Make the new user.
	l.s.u = core.NewUser("oddcomm/src/ts6", &(l.s), true,
		string(params[3]), nil)

	if l.s.u == nil {
		// SID in use!
		l.c.Close()
		return
	}
}


// Server registration command.
func cmdServer(source interface{}, params [][]byte) {
}
