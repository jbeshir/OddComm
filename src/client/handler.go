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
			_, command, params, perr := irc.Parse(Commands, line,
				c.u.Registered())

			// If we successfully got a command, run it.
			if command != nil {

				// If it's an oper command, check permissions.
				if command.OperFlag != "" && !perm.HasOperCommand(c.u, command.OperFlag, command.Name) {
					c.SendLineTo(nil, "481", ":You do not have the appropriate privileges to use this command.")
				} else {
					command.Handler(c, params)
				}
			} else if perr != nil {

				// The IRC protocol is stupid.
				switch perr.Num {
				case irc.CmdNotFound:
					if c.u.Registered() {
						c.SendLineTo(nil, "421", "%s :%s", perr.CmdName, perr)
					}
				case irc.CmdForRegistered:
					c.SendFrom(nil, "451 %s :%s",
						perr.CmdName, perr)
				case irc.CmdForUnregistered:
					c.SendFrom(nil, "462 %s :%s",
						c.u.Nick(), perr)
				default:
					c.SendFrom(nil, "461 %s %s :%s",
						c.u.Nick(), perr.CmdName,
						perr)
				}
			}

			// If we have remaining input for the next line, move
			// it down and cut the buffer to it.
			// Otherwise, clear it.
			if len(b)-eol-1 >= 0 {
				copy(b, b[eol+1:])
				b = b[:len(b)-eol-1]
			} else {
				b = b[:0]
				break
			}
		}

		count = len(b)
	}
}


// Output goroutine function for a client.
// Used when the client is switched to buffered writes.
// This happens in two cases: When writing output blocks for too long, or a
// goroutine wishes to do blocking writes over time (in which case it "becomes"
// the output goroutine for the duration).
//
// While existing, it owns writing to the socket, and the holder of the
// client's mutex may only append to the output buffer. The only permitted way
// for any other changes, including deletion, to the output buffer to be made,
// is for this goroutine to do it while holding the mutex, until this goroutine
// terminates.
//
// If f is not nil, it will be called as often as the connection can support,
// and its result written to the client, until it returns a nil slice, at which
// time this function will spawn a new output goroutine if needed and return.
// This call will replace an existing output goroutine if necessary.
//
// If f is nil, we will assume we are being called either by a previous output
// goroutine, or a goroutine which has just turned buffering on, and do not
// need to replace an existing goroutine.
func output(c *Client, f func() []byte) {
	var n int
	var err os.Error

	// We hold the client mutex whenever not writing.
	c.mutex.Lock()

	// If we're replacing an existing output goroutine,
	// add to the count and wait.
	if f != nil {
		if c.outbuf != nil {
			c.outcount++
			c.mutex.Unlock()
			<-c.outchan
			c.mutex.Lock()
		} else {
			c.bufferOn()
		}
	}

	// Get the current length of the output buffer.
	n = len(c.outbuf)

	// While we have something to write...
	for {
		// Unlock the client mutex.
		c.mutex.Unlock()

		// Write from the output buffer, if we have output.
		if n > 0 {
			_, err = c.conn.Write(c.outbuf[:n])
		}

		// If f != nil and the buffer was below half full,
		// call f and write its result.
		if f != nil && err == nil && n <= cap(c.outbuf)/2 {
			buf := f()
			if buf == nil {
				// We're done. If there's a waiting output
				// goroutine, tell them to go. If not and the
				// output buffer is not empty, spawn a new
				/// output goroutine.
				c.mutex.Lock()
				if c.outcount != 0 {
					c.outcount--
					c.outchan <- true
				} else if len(c.outbuf) != 0 {
					go output(c, nil)
				} else {
					c.bufferOff()
				}
				break
			}

			// Write the result.
			_, err = c.conn.Write(buf)
		}

		// Relock the mutex.
		c.mutex.Lock()

		// If writing failed, delete the user, suppressing writes.
		if err != nil {
			c.bufferOff()
			c.disconnecting |= 2
			c.delete(err.String())
			break
		}

		// If f is nil and we've been asked to stop, do so.
		// Otherwise, they have to wait until f is done.
		if f == nil {
			if c.outcount != 0 {
				c.outcount--
				c.outchan <- true
				break
			}
		}

		// Get more to write.
		// If we run out and f is nil, turn off buffered I/O.
		if len(c.outbuf) == n && f == nil {
			c.bufferOff()
			break
		} else {
			copy(c.outbuf[:len(c.outbuf)-n], c.outbuf[n:])
			c.outbuf = c.outbuf[:len(c.outbuf)-n]
			n = len(c.outbuf)
		}
	}

	c.mutex.Unlock()
}
