package client

import "fmt"
import "net"
import "os"
import "sync"

import "oddcomm/src/core"


// Handle a client connection.
type Client struct {
	mutex         sync.Mutex
	conn          *net.TCPConn
	u             *core.User
	outbuf        []byte
	outcount      int
	outchan       chan bool
	disconnecting uint8
	nicked        bool
}

// User returns this client's user.
func (c *Client) User() *core.User {
	return c.u
}

// Disconnects the client with the given message. This internal method assumes
// it is being called with the client mutex already held.
//
// Should be called once to perform the disconnection, and then again when the
// the output buffer is empty, and when the input goroutine has terminated,
// if these were not true originally. The message is ignored in later calls,
// and the user is only deleted once.
// 
// The client is fully deleted when this method is called when the output
// buffer is empty, and the input goroutine has terminated.
func (c *Client) delete(message string) {

	// First deletion only stuff.
	if c.disconnecting&1 == 0 {

		// Mark us as disconnecting.
		c.disconnecting |= 1

		// If the user has not already been deleted, delete it.
		c.u.Delete(nil, message)
	}

	// Send them a goodbye message if we've not already sent one.
	// This also suppresses further writes to the client, and this bit can
	// be set prior to calling delete() if a quit message should not be
	// sent, such as in the case of a write error.
	if c.disconnecting&2 == 0 {
		username := c.u.Data("ident")
		if username == "" {
			username = "unknown"
		}
		c.write([]byte(fmt.Sprintf("ERROR :Closing link: (%s@%s) [%s]\r\n", username, c.u.Data("hostname"), message)))
		c.disconnecting &= 2
	}

	// If the output buffer is done...
	if c.outbuf == nil {
		// If we haven't already closed the connection, do so.
		// This will cause the input goroutine to terminate if it
		// has not yet done so.
		if c.disconnecting&4 == 0 {
			c.conn.Close()
			c.disconnecting |= 4
		}

		// If the input goroutine has terminated, fully delete the
		// client if we haven't already.
		if c.disconnecting&8 != 0 && c.disconnecting&16 == 0 {
			decClient(c)
			c.disconnecting |= 16
		}
	}
}

// Write a raw line to the client. This internal method assumes it is being
// called with the client mutex already held.
func (c *Client) write(line []byte) bool {

	// If the client is disconnecting, drop all writes to it.
	if c.disconnecting&2 != 0 {
		return false
	}

	// Define function to append to the output buffer.
	var appendfunc = func(line []byte) bool {

		// If we've overflowed our output buffer, kill the client.
		if cap(c.outbuf)-len(c.outbuf) < len(line) {
			// Suppress output prior to calling delete, so it
			// does not attempt to send a quit message.
			c.disconnecting |= 2
			c.delete("Output Buffer Exceeded")
			return false
		}

		// Otherwise, append this to it.
		start := len(c.outbuf)
		c.outbuf = c.outbuf[:start+len(line)]
		copy(c.outbuf[start:], line)
		return true
	}

	// If we're not using buffered output, try to write it directly.
	if c.outbuf == nil {

		// Try to write.
		n, err := c.conn.Write(line)
		if err != nil && err.(net.Error).Timeout() == false {
			// Suppress output prior to calling delete, so it
			// does not attempt to send a quit message.
			c.disconnecting |= 2
			c.delete(err.String())
			return false
		}

		// If it takes too long or we can't write it all, turn on
		// buffered I/O, start a goroutine, and tell it to go.
		if n != len(line) {
			c.bufferOn()
			if !appendfunc(line[n:]) {
				return false
			}
			go output(c, nil)
		}
	} else {
		// If we're using buffered output, add it to the buffer.
		if !appendfunc(line) {
			return false
		}
	}

	return true
}

// Switch the client to buffered I/O. The caller of this becomes the output
// goroutine and is responsible for ensuring that output happens.
// Assumes the caller holds the client mutex.
func (c *Client) bufferOn() {
	c.outbuf = make([]byte, 0, 8192)
	c.conn.SetWriteTimeout(0)
}

// Switch the client from buffered I/O. This can only be called from an output
// goroutine which is terminating, and holds the client mutex, with no other
// output goroutine waiting to run.
func (c *Client) bufferOff() {
	c.outbuf = nil
	c.outchan = nil
	if c.disconnecting&2 != 0 {
		c.delete("Output Done")
	} else {
		c.conn.SetWriteTimeout(1000)
	}
}

// Write a raw line to the client.
func (c *Client) Write(line []byte) (int, os.Error) {
	var written bool

	c.mutex.Lock()
	written = c.write(line)
	c.mutex.Unlock()

	if !written {
		return 0, os.NewError("Client disconnecting.")
	}
	return len(line), nil
}

// Write a formatted line from the given user, addressed to this client.
// u may be nil, in which case, the line will be from the server.
// A line ending will be automatically appended.
func (c *Client) WriteTo(u *core.User, cmd string, format string, args ...interface{}) {
	if u != nil {
		fmt.Fprintf(c, ":%s!%s@%s %s %s %s\r\n", u.Nick(),
			u.GetIdent(), u.GetHostname(), cmd,
			c.u.Nick(), fmt.Sprintf(format, args...))
	} else {
		fmt.Fprintf(c, ":%s %s %s %s\r\n", "Server.name", cmd,
			c.u.Nick(), fmt.Sprintf(format, args...))
	}
}

// Write the given line, prefixed by the given source.
// u may be nil, in which case, the line will be from the server.
// A line ending will be automatically appended.
func (c *Client) WriteFrom(u *core.User, format string, args ...interface{}) {
	if u != nil {
		fmt.Fprintf(c, ":%s!%s@%s %s\r\n", u.Nick(),
			u.GetIdent(), u.GetHostname(),
			fmt.Sprintf(format, args...))
	} else {
		fmt.Fprintf(c, ":%s %s\r\n", "Server.name",
			fmt.Sprintf(format, args...))
	}
}

// WriteBlock permits blocking writes to the client. It calls the given
// function repeatedly to generate output until it returns nil, writing it over
// time. In general, the only place blocking is permitted is in a command
// handler for that client.
func (c *Client) WriteBlock(f func() []byte) {
	output(c, f)
}
