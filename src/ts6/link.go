package ts6

import "bytes"
import "fmt"
import "net"
import "time"

import "oddcomm/src/core"
import "oddcomm/lib/irc"


// Handle a single (potential) server link.
// outgoing indicates whether it is outgoing or incoming.
func link(c *net.TCPConn, outgoing bool) {
	errMsg := "Input Error"

	serverMutex.Lock()

	// Create a new server, and add it to the server list.
	l := new(local)
	l.local = l
	l.next = servers
	if servers != nil {
		servers.prev = &(l.server)
	}
	servers = &(l.server)
	l.c = c

	serverMutex.Unlock()

	// Defer deletion of the server. If it's already deleted, no harm done.
	defer func() {
		l.Delete(errMsg)
	}()

	if outgoing {
		link_auth(l)
	}

	errMsg = irc.ReadLine(l.c,  make([]byte, 20960), func(line []byte) {

		// Parse the line.
		prefix, command, params, perr := irc.Parse(commands, line, l.authed)

		// Look up the server or user this command is from.
		var source interface{}
		if len(prefix) == 9 {
			u := core.GetUser(string(prefix));
			if u == nil {
				source = nil
			} else if u.Owner() != me {
				source = nil
			} else {
				userver := u.Owndata().(*server)
				if userver.local == l {
					source = u
				}
			}
		} else if len(prefix) == 3 {
			v := core.GetSID(string(prefix))
			if v == nil {
				source = nil
			} else if s, ok := v.(*server); ok {
				if s.local == l {
					source = s
				}
			}
		} else if len(prefix) == 0 {
			// No prefix; it's from this server.
			source = &(l.server)
		} else {
			// Prefix is gibberish.
			source = nil
		}

		// If we successfully got a command and source, run it.
		if source != nil && command != nil {
			command.Handler(source, params)
		} else if perr != nil {

			// The IRC protocol is stupid.
			switch perr.Num {
			case irc.CmdNotFound:
				irc.SendLine(l, from(nil), l.sid, "421", "%s :%s",
					perr.CmdName, perr)
			case irc.CmdForRegistered:
				irc.SendFrom(l, from(nil), "451 %s :%s", perr.CmdName,
					perr)
			case irc.CmdForUnregistered:
				irc.SendFrom(l, from(nil), "462 %s :%s", l.sid, perr)
			default:
				irc.SendFrom(l, from(nil), "461 %s %s :%s", l.sid,
					perr.CmdName, perr)
			}
		}
	})
}


// Auth to a locally linked server.
func link_auth(l *local) {

	// No configuration!
	fmt.Fprintf(l, "PASS supertest TS 6 :1AA\r\nCAPAB :QS ENCAP\r\nSERVER Test.net 1 :Testing\r\nSVINFO 6 6 0 :%d\r\n", time.Seconds())
	l.auth_sent = true
}

// Burst to a locally linked server.
func link_burst(l *local) {
	core.Sync(
		func(u *core.User, hook bool) {
			if !hook {
				send_uid(l, u)
			}
		},
		func(ch *core.Channel, hook bool) {
			b := bytes.NewBuffer(make([]byte, 0, 512))
			user := bytes.NewBuffer(make([]byte, 0, 50))
			if !hook {
				it := ch.Users()
				for {
					fmt.Fprintf(b, "SJOIN %d #%s +n :", 0, ch.Name())
					for ;it != nil; it = it.ChanNext() {
						user.WriteString(it.User().ID())
						user.Write([]byte(" "))
						if b.Len() + user.Len() > 509 {
							break
						}
						user.WriteTo(b)
						user.Reset()
					}
					b.Write([]byte("\r\n"))
					b.WriteTo(l)
					b.Reset()

					if it == nil {
						break
					}
				}
			}
		},
		func(hook bool) {
			// We currently *assume* this is called after all the other
			// hooks.
			if !hook {
				l.burst_sent = true
			}
		})
}


// Introduce a user to a given local server.
func send_uid(l *local, u *core.User) {
	irc.SendFrom(l, from(nil), "UID %s 1 %d +i %s %s %s %s :%s", u.Nick(),
		u.NickTS(), u.GetIdent(), u.GetHostname(), u.GetIP(), u.ID(),
		u.Data("realname"))
}

// Iterates all servers, running the given function on each.
// May omit deleted servers, who no longer are cared about.
// May omit servers added after the start of the call, who presumably do not need to
// know about whatever event already happened to result in this call.
func all(f func(l *local)) {
	serverMutex.Lock()
	for s := servers; s != nil; s = s.next {
		serverMutex.Unlock()
		f(s.local)
		serverMutex.Lock()
	}
	serverMutex.Unlock()
}
