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
	disconnecting uint8
	nicked        bool
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

		// Remove this client from our live client lists.
		killClient(c)

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
			delClient(c)
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
			// Suppress output prior to calling delete, so it
			// does not attempt to send a quit message.
			c.disconnecting |= 2
			c.delete(err.String())
			return false
		}

		// If it takes too long or we can't write it all, make an
		// output buffer and switch to buffered I/O.
		if n != len(line) {
			c.outbuf = make([]byte, 0, 4096)
			if !appendfunc(line[n:]) {
				return false
			}
			c.conn.SetWriteTimeout(0)
			go output(c, len(line)-n)
		}
	} else {
		// If we're using buffered output, add it to the buffer.
		if !appendfunc(line) {
			return false
		}
	}

	return true
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
