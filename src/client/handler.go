package client

import "os"

import "oddcomm/lib/irc"
import "oddcomm/lib/perm"


// Input goroutine function for a client.
// Own readings on the socket.
func input(c *Client) {

	// Stop panics here, so they only disconnect the client they affect.
	//defer func() {
	//	recover()
	//}()

	// Delete the client when this returns.
	errMsg := "Input Error"
	defer func() {
		c.mutex.Lock()
		c.disconnecting |= 4 // Marks input as done.
		c.delete(errMsg)
		c.mutex.Unlock()
	}()

	b := make([]byte, 2096)
	var count int
	for {
		// If we have no room in our input buffer to read, the user
		// has overrun their input buffer.
		if count == cap(b) {
			errMsg = "Input Buffer Exceeded"
			break
		}

		// Try to read from the user.
		n, err := c.conn.Read(b[count:cap(b)])
		if err != nil {
			// This happens if the user is disconnected
			// by other code. In this case, the error message
			// will be ignored.
			errMsg = err.String()
			break
		}
		count += n
		b = b[0:count]

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
			line := b[0:eol]
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
				line = line[0:end]
			}

			// Ignore blank lines.
			if len(line) == 0 {
				if len(b)-eol-1 >= 0 {
					b = b[0 : len(b)-eol-1]
					continue
				} else {
					b = b[0:0]
					break
				}
			}

			// Parse the line, ignoring any specified origin.
			_, command, params, perr := irc.Parse(Commands, line,
				c.u.Registered())

			// If we successfully got a command, run it.
			if command != nil {

				// If it's an oper command, check permissions.
				if command.OperFlag != "" && !perm.HasOperCommand(c.u, command.OperFlag, command.Name) {
					c.WriteTo(nil, "481", ":You do not have the appropriate privileges to use this command.")
				} else {
					command.Handler(c.u, c, params)
				}
			} else if perr != nil {

				// The IRC protocol is stupid.
				switch perr.Num {
				case irc.CmdNotFound:
					if c.u.Registered() {
						c.WriteTo(nil, "421", "%s :%s", perr.CmdName, perr)
					}
				case irc.CmdForRegistered:
					c.WriteFrom(nil, "451 %s :%s",
						perr.CmdName, perr)
				case irc.CmdForUnregistered:
					c.WriteFrom(nil, "462 %s :%s",
						c.u.Nick(), perr)
				default:
					c.WriteFrom(nil, "461 %s %s :%s",
						c.u.Nick(), perr.CmdName,
						perr)
				}
			}

			// If we have remaining input for the next line, move
			// it down and cut the buffer to it.
			// Otherwise, clear it.
			if len(b)-eol-1 >= 0 {
				for i := 0; i < len(b)-eol-1; i++ {
					b[i] = b[eol+1+i]
				}
				b = b[0 : len(b)-eol-1]
			} else {
				b = b[0:0]
				break
			}
		}

		count = len(b)
	}
}


// Output goroutine for a client.
// Only spawned when writing output blocks for too long, and the client is
// switched over to buffered output. While existing, it owns writing to the
// socket, and the holder of the client's mutex may only append to the output
// buffer. The only permitted way for any other changes, including deletion,
// to the output buffer to be made, is for this goroutine to do it while
// holding the mutex, until this goroutine terminates.
func output(c *Client, n int) {

	// While we have something to write...
	for n > 0 {

		// Write it.
		var err os.Error
		n, err = c.conn.Write(c.outbuf[0:n])

		// If writing failed, delete the user- but close the connection
		// first to suppress write attempts.
		if err != nil {
			c.mutex.Lock()
			c.outbuf = nil
			c.conn.Close()
			c.delete(err.String())
			c.mutex.Unlock()
			break
		}

		// Get more to write.
		// If we run out, turn off buffered I/O.
		c.mutex.Lock()
		if len(c.outbuf) == n {
			c.outbuf = nil
			if c.disconnecting & 2 != 0 {
				c.delete("Output Done")
			} else {
				c.conn.SetWriteTimeout(1000)
			}
			n = 0
		} else {
			for i := 0; i < len(c.outbuf)-n; i++ {
				c.outbuf[i] = c.outbuf[n+i]
			}
			c.outbuf = c.outbuf[0 : len(c.outbuf)-n]
			n = len(c.outbuf)
		}
		c.mutex.Unlock()
	}
}
