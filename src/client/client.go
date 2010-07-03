package client

import "fmt"
import "net"
import "os"

import "oddircd/core"


// Handle a client connection.
// u != nil - The client is registered.
// unreg != nil - The client is unregistered.
// u == nil && unreg == nil - The client is disconnecting.
type Client struct {
	cchan     chan clientRequest
	conn      net.Conn
	u         *core.User
	unreg     *unregUser
	inputDone bool
	outbuf    []byte
}

// Stores information for an unregistered user.
type unregUser struct {
	nick string
	meta map[string]string
}


// Write a raw line to the client. This internal method assumes it is being
// called from the client goroutine.
func (c *Client) write(line []byte) {

	// Define function to append to the output buffer.
	var appendfunc = func(line []byte) bool {

		// If we've overflowed our output buffer, kill the client.
		if cap(c.outbuf)-len(c.outbuf) < len(line) {
			c.u = nil
			c.unreg = nil
			return false
		}

		// Otherwise, append this to it.
		start := len(c.outbuf)
		c.outbuf = c.outbuf[0 : start+len(line)]
		for i := 0; i < len(line); i++ {
			c.outbuf[start+i] = line[i]
		}
		return true
	}

	// If we're not using buffered output, try to write it directly.
	if c.outbuf == nil {

		// Try to write.
		n, err := c.conn.Write(line)
		if err != nil && err.(net.Error).Timeout() == false {
			c.u = nil
			c.unreg = nil
			return
		}

		// If it takes too long or we can't write it all, make an
		// output buffer and switch to buffered I/O.
		if n != len(line) {
			c.outbuf = make([]byte, 4096)
			c.outbuf = c.outbuf[0:0]
			if !appendfunc(line[n:len(line)]) {
				return
			}
			c.conn.SetWriteTimeout(0)
			go output(c, len(line)-n)
		}
	} else {
		// If we're using buffered output, add it to the buffer.
		if !appendfunc(line) {
			return
		}
	}
}

// Write a raw line to the client.
func (c *Client) Write(line []byte) (int, os.Error) {

	written := makeRequest(c, func() {
		c.write(line)
	})

	if !written {
		return 0, os.NewError("Client disconnecting.")
	}

	return len(line), nil
}

// Quit the client.
// Message is written to them on a line of its own first, if non-null.
func (c *Client) Quit(message []byte) {
	makeRequest(c, func() {
		c.u = nil
		c.unreg = nil
		if message != nil {
			c.write(message)
			c.write([]byte("\r\n"))
		}
	})
}

// Privmsg sends a PRIVMSG to the client.
func (c *Client) Privmsg(source string, message []byte) {
	fmt.Fprintf(c, "%s PRIVMSG Namegduf :%s", source, message)
}
