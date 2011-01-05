package ts6

import "fmt"
import "net"
import "time"

import "oddcomm/lib/irc"


// Handle a single (potential) server link.
// outgoing indicates whether it is outgoing or incoming.
func link(c *net.TCPConn, outgoing bool) {
	var errMsg string
	_ = errMsg

	server := new(local)
	server.c = c
	if outgoing {
		link_auth(server)
	}

	b := make([]byte, 20960)
	var count int
	for {
		// If we have no room in our input buffer to read, the user
		// has overrun their input buffer.
		if count == cap(b) {
			errMsg = "Input Buffer Exceeded"
			break
		}

		// Try to read from the user.
		n, err := server.c.Read(b[count:cap(b)])
		if err != nil {
			// This happens if the user is disconnected
			// by other code. In this case, the error message
			// will be ignored.
			errMsg = err.String()
			break
		}
		count += n
		b = b[:count]

		for {
			// Search for an end of line, then keep going until we
			// stop finding eol characters, to eat as many as
			// possible in the same operation.
			eol := -1
			for i := range b {
				if b[i] == '\r' || b[i] == '\n' || b[i] == 0 {
					eol = i
				} else if eol != -1 {
					break
				}
			}

			// If we didn't find one, wait for more input.
			if eol == -1 {
				break
			}

			// Get the line, with no line endings.
			line := b[:eol]
			end := len(line)
			for end > 0 {
				endchar := line[end-1]
				if endchar == '\r' || endchar == '\n' {
					end--
				} else {
					break
				}
			}
			if end != len(line) {
				line = line[:end]
			}

			// Ignore blank lines.
			if len(line) == 0 {
				if len(b)-eol-1 >= 0 {
					b = b[:len(b)-eol-1]
					continue
				} else {
					b = b[:0]
					break
				}
			}

			// Parse the line, ignoring any specified origin.
			prefix, command, params, perr := irc.Parse(commands,
				line, true)

			// Look up the user this command is from.
			_ = prefix

			// If we successfully got a command, run it.
			if command != nil {
				command.Handler(server, params)
			} else if perr != nil {

				// The IRC protocol is stupid.
				switch perr.Num {
				case irc.CmdNotFound:
				//	server.c.WriteTo(nil, "421", "%s :%s",
				//		perr.CmdName, perr)
				case irc.CmdForRegistered:
				//	server.c.WriteFrom(nil, "451 %s :%s",
				//		perr.CmdName, perr)
				case irc.CmdForUnregistered:
				//	server.c.WriteFrom(nil, "462 %s :%s",
				//		c.u.Nick(), perr)
				default:
				//	server.c.WriteFrom(nil, "461 %s %s :%s",
				//		c.u.Nick(), perr.CmdName,
				//		perr)
				}
				fmt.Fprintf(c, "421 %s :%s\n", perr.CmdName,
					perr)
			}

			// If we have remaining input for the next line, move
			// it down and cut the buffer to it.
			// Otherwise, clear it.
			if len(b)-eol-1 >= 0 {
				for i := 0; i < len(b)-eol-1; i++ {
					b[i] = b[eol+1+i]
				}
				b = b[:len(b)-eol-1]
			} else {
				b = b[:0]
				break
			}
		}

		count = len(b)
	}
}


// Auth to a locally linked server.
func link_auth(server *local) {

	// No configuration!
	server.c.Write([]byte("PASS ultratest TS 6 :0ZZ\n"))
	server.c.Write([]byte("SERVER Test.net 1 :Testing\n"))
	fmt.Fprintf(server.c, "SVINFO 6 6 0 :%d\n", time.Seconds())
}

// Burst to a locally linked server.
func link_burst(server *local) {
}
