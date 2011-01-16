package ts6

import "fmt"

import "oddcomm/src/core"
import "oddcomm/lib/irc"


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

	// Add capab command.
	c = new(irc.Command)
	c.Name = "CAPAB"
	c.Handler = cmdCapab
	c.Minargs = 1
	c.Maxargs = 1
	c.Unregged = 2
	commands.Add(c)

	// Add server command.
	c = new(irc.Command)
	c.Name = "SERVER"
	c.Handler = cmdServer
	c.Minargs = 3
	c.Maxargs = 3
	c.Unregged = 1
	commands.Add(c)

	// Add SVINFO command.
	c = new(irc.Command)
	c.Name = "SVINFO"
	c.Handler = cmdSvinfo
	c.Minargs = 4
	c.Maxargs = 4
	c.Unregged = 2
	commands.Add(c)

	// Add PING command.
	c = new(irc.Command)
	c.Name = "PING"
	c.Handler = cmdPing
	c.Minargs = 1
	c.Maxargs = 1
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
	if &(l.server) != s {
		return
	}

	// If the remote server isn't speaking TS6, drop it.
	if string(params[2]) != "6" {
		l.c.Close()
		return
	}

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

	// Set the password.
	l.pass = string(params[0])

	// Set the SID, if there's no conflict.
	if core.NewSID(sid, &(l.server)) {
		if l.server.sid != "" {
			core.ReleaseSID(l.server.sid)
		}
		l.server.sid = sid
	} else {
		return
	}
}


// CAPAB command.
// Provides us with the server's capabilities.
// The connection uses the intersection of these and those we support.
func cmdCapab(source interface{}, params [][]byte) {
	s, ok := source.(*server)
	if !ok {
		return
	}

	// If this isn't a local server, ignore it.
	l := s.local
	if &(l.server) != s {
		return
	}

	// We just ignore this for now; we only have two capabilities and
	// they are required by TS6, so we can assume they're being used.
}


// Server adding command.
// May only be from a server.
func cmdServer(source interface{}, params [][]byte) {
	s, ok := source.(*server)
	if !ok {
		return
	}

	// Adding remote servers is not yet implemented.
	// If this isn't a local server, ignore it.
	l := s.local
	if &(l.server) != s {
		return
	}

	// Set the server name.
	s.name = string(params[0])

	// Set the server description.
	s.desc = string(params[2])
}

// Server information command.
// Provides TS version and current time, for safety checks.
// Final step in registering a local server.
func cmdSvinfo(source interface{}, params [][]byte) {
	s, ok := source.(*server)
	if !ok {
		return
	}

	// If this isn't a local server, ignore it.
	l := s.local
	if &(l.server) != s {
		fmt.Printf("%v %v", &(l.server), l)
		return
	}

	// Check the given TS version is 6.
	if string(params[0]) != "6" {
		l.c.Close()
		return
	}

	// Registration steps complete.
	// Try to authenticate them.
	// Blank this out for now.
	if s.name != "" && s.desc != "" && l.pass != "" {
		// It's authed.
		l.authed = true

		// Auth and burst to them.
		if !l.auth_sent {
			link_auth(l)
		}
		link_burst(l)
	} else {
		l.c.Close()
	}
}


// Add PING command.
// Also used by the remote server to detect end of burst.
func cmdPing(source interface{}, params[][]byte) {
	s, ok := source.(*server)
	if !ok {
		return
	}

	// Echo it back, including the remote destination if it has one.
	if len(params) == 1 {
		fmt.Fprintf(s.local, "1AA PONG %s\n", params[0])
	} else {
		fmt.Fprintf(s.local, "1AA PONG %s %s\n", params[0], params[1])
	}
}