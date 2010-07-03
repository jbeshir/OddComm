package client

import "bytes"
import "fmt"
import "os"

import "oddircd/core"
import "oddircd/irc"


// Primary goroutine function for a client.
// Owns the client, shares the socket; must not read from it, nor set it nil.
// Shares the output buffer if one exists; see the output goroutine's comment
// for more details on the rules for that.
func clientHandler(c *Client) {

	// Stop panics here, so they only disconnect the client they affect.
	defer func() {
		recover()
	}()

	// Declare shutdown variables.
	// Shutdown is gradual; we wait for I/O to complete first.
	var (
		closing    bool
		sockclosed bool
		chanclosed bool
	)

	// Add us to the client goroutine list, and remove us on returning.
	addClient(c)
	defer func() { delClient(c) }()

	// Close the connection on unclean shutdown.
	defer c.conn.Close()

	// Spawn the client input goroutine.
	go input(c)

	// Handle requests until they stop coming.
	for request := range c.cchan {

		// Read them, and run their function.
		request.f()
		request.done <- true

		// If we just got told to close down, mark our client as dead
		// and note that the client is being disconnected. This means
		// new requests will only be arriving from the input and
		// output goroutines.
		if !closing && c.u == nil && c.unreg == nil {
			killClient(c)
			closing = true
		}

		// If disconnecting, wait until the output goroutine has gone,
		// then close the socket, stopping the input goroutine.
		if closing && !sockclosed && c.outbuf == nil {
			c.conn.Close()
			sockclosed = true
		}

		// If disconnecting, and we've been told the input goroutine
		// has finished dying, we will no longer be receiving new
		// requests; there are no writers. Close our channel.
		// We will terminate when all messages are handled.
		if sockclosed && !chanclosed && c.inputDone {
			close(c.cchan)
			chanclosed = true
		}
	}
}


// Input goroutine function for a client.
// Does NOT own the client and must ask the output goroutine for most
// information and all mutations. Does own reading on the socket.
func input(c *Client) {

	// Stop panics here, so they only disconnect the client they affect.
	defer func() {
		recover()
	}()

	// Defer marking ourselves as done, so the client goroutine can
	// terminate. We also disconnect the client if input fails early.
	defer func() {
		makeDirectRequest(c, func() {
			c.u = nil
			c.unreg = nil
			c.inputDone = true
		})
	}()

	b := make([]byte, 1024)
	var count int
	var clipline bool
	for {
		n, err := c.conn.Read(b[count:1024])
		if err != nil {
			makeDirectRequest(c, func() {
				c.u = nil
				c.unreg = nil
			})
			break
		}
		count += n
		b = b[0:count]

		for {
			// If we're still waiting for the end of a clipped
			// line, see if we've reached it.
			if clipline {

				// Find the end of the first set of EOL
				// characters, eating as many as possible.
				eol := -1
				for i := range b {
					if b[i] == '\r' || b[i] == '\n' {
						eol = i
					} else if eol != -1 {
						break
					}
				}
				if eol != -1 {
					clipline = false
				}

				// If we reached the end of the line to clip
				// and had remaining input, move it down.
				// Otherwise, clear the buffer and break.
				if eol != -1 && len(b)-eol-1 >= 0 {
					for i := 0; i < len(b)-eol-1; i++ {
						b[i] = b[eol+1+i]
					}
					b = b[0 : len(b)-eol-1]
				} else {
					b = b[0:0]
					break
				}

				continue
			}

			// Search for an end of line, then keep going until we
			// stop finding eol characters, to eat as many as
			// possible in the same operation. If the line is more
			// than 512 bytes long, clip the line there.
			eol := -1
			for i := range b {
				if b[i] == '\r' || b[i] == '\n' {
					eol = i
				} else if eol != -1 {
					break
				} else if i >= 512 {
					clipline = true
					eol = i
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
			_, command, params := irc.Parse(line)

			// Command dispatch!
			// Kinda.
			// Well, no, not really. Not yet.
			if bytes.Equal(command, []byte("NICK")) &&
				len(params) > 0 {
				fmt.Fprintf(c, "Hi %s!\r\n", params[0])
			} else if bytes.Equal(command, []byte("QUIT")) {
				// A perfectly valid implementation.
				// It DOES disconnect the user.
				core.Shutdown()
			} else {
				// Polly wants a cracker.
				c.Write(line)
				c.Write([]byte("\r\n"))
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
// socket, and the client goroutine may only append to the output buffer. The
// only permitted way for any other changes, including deletion, to the output
// buffer to be made, is blocking this goroutine while the client goroutine
// does it.
func output(c *Client, n int) {

	// While we have something to write...
	for n > 0 {

		// Write it.
		var err os.Error
		n, err = c.conn.Write(c.outbuf[0:n])

		// If writing failed, murder the user and run away.
		if err != nil {
			c.conn.Close()
			break
		}

		// Get more to write.
		// If we run out, turn off buffered I/O.
		makeDirectRequest(c, func() {
			if len(c.outbuf) == n {
				c.outbuf = nil
				c.conn.SetWriteTimeout(1000)
				n = 0
			} else {
				for i := 0; i < len(c.outbuf)-n; i++ {
					c.outbuf[i] = c.outbuf[n+i]
				}
				c.outbuf = c.outbuf[0 : len(c.outbuf)-n]
				n = len(c.outbuf)
			}
		})
	}
}
