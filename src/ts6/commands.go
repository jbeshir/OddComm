package ts6

import "oddcomm/lib/irc"


// Commands added here will be called with either a server or a user.
var commands irc.CommandDispatcher = irc.NewCommandDispatcher()


func init() {
	var c *irc.Command

	// Add pass command.
	c = new(irc.Command)
	c.Name = "PASS"
	c.Handler = cmdPass
	c.Minargs = 4
	c.Maxargs = 4
	c.Unregged = 2
	commands.Add(c)

	// Add pass command.
	c = new(irc.Command)
	c.Name = "SERVER"
	c.Handler = cmdServer
	c.Minargs = 3
	c.Maxargs = 3
	c.Unregged = 2
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
	if &l.server != s {
		return
	}

	// Set the password.
	l.pass = string(params[0])

	// Validate the given SID.
	sid := string(params[3])
	if len(sid) != 3 {
		return
	}
	if sid[0] < '0' || sid[0] > '9' {
		return
	}
	if (sid[1] < '0' || sid[1] > '9') && (sid[1] < 'A' || sid[1] > 'Z') {
		return
	}
	if (sid[2] < '0' || sid[2] > '9') && (sid[2] < 'A' || sid[2] > 'Z') {
		return
	}

	// Set the SID, if there's no conflict.
	if core.NewSID(sid, &l.server) {
		l.server.sid = sid
	} else {
		return
	}
}


// Server registration command.
// Either registers a local server, or adds a new remote server.
// May only be from a server.
func cmdServer(source interface{}, params [][]byte) {
	s, ok := source.(*server)
	if !ok {
		return
	}

	// Adding remote servers is not yet implemented.
	// If this isn't a local server, ignore it.
	l := s.local
	if &l.server != s {
		return
	}
}
