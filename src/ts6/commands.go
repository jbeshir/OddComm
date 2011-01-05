package ts6

import "io"

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


// Password authentication. Does server user creation.
// Ignore the user here; specifying one is invalid. We care about the server.
func cmdPass(u *core.User, w io.Writer, params [][]byte) {
	l := w.(*local)

	// Ignore PASS commands from already authed servers.
	// They can happen if they have an invalid source.
	if l.authed {
		return false
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
// If the user here is nil, it's a local server we need to do authentication
// for, which is the writer.
// If it is not, it is an unregistered remote server created for this command,
// or for some reason a remote server sent a server command from a user.
func cmdServer(u *core.User, w io.Writer, params [][]byte) {
}
