package client

import "fmt"
import "net"
import "os"

import "oddcomm/src/core"


// Handle a client connection.
type Client struct {
	cchan     chan clientRequest
	conn      *net.TCPConn
	u         *core.User
	disconnecting bool
	inputDone bool
	outbuf    []byte
}

// Disconnects the client with the given message. This internal method assumes
// it is being called from the client goroutine.
func (c *Client) delete(message string) {

	// Remove this client from our live client lists.
	killClient(c)

	// Send them a goodbye message.
	username := c.u.Data("ident")
	if username == "" {
		username = "unknown"
	}
	c.write([]byte(fmt.Sprintf("ERROR :Closing link: (%s@%s) [%s]\r\n", username, c.u.Data("hostname"), message)))

	// If the user has not already been deleted, delete it.
	c.u.Delete(nil, message)

	c.disconnecting = true
}

// Write a raw line to the client. This internal method assumes it is being
// called from the client goroutine.
func (c *Client) write(line []byte) {
	
	// If the client is disconnecting, drop all writes to it.
	if (c.disconnecting) {
		return
	}

	// Define function to append to the output buffer.
	var appendfunc = func(line []byte) bool {

		// If we've overflowed our output buffer, kill the client.
		if cap(c.outbuf)-len(c.outbuf) < len(line) {
			// Mark it as disconnecting before calling the delete
			// method to suppress the quit message.
			c.disconnecting = true
			c.delete("Output Buffer Exceeded")
			c.outbuf = nil
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
			// Mark it as disconnecting before calling the delete
			// method to suppress the quit message.
			c.disconnecting = true
			c.delete(err.String())
			c.outbuf = nil
			return
		}

		// If it takes too long or we can't write it all, make an
		// output buffer and switch to buffered I/O.
		if n != len(line) {
			c.outbuf = make([]byte, 0, 4096)
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

// Write a formatted line from the given user, addressed to this client.
// u may be nil, in which case, the line will be from the server.
// A line ending will be automatically appended.
func (c *Client) WriteTo(u *core.User, cmd string, format string,
                           args ...interface{}) {
	if u != nil {
		fmt.Fprintf(c, ":%s!%s@%s %s %s %s\r\n", u.Nick(),
		            u.GetIdent(), u.GetHostname(), cmd,
		            c.u.Nick(), fmt.Sprintf(format, args))
	} else {
		fmt.Fprintf(c, ":%s %s %s %s\r\n", "Server.name", cmd,
		            c.u.Nick(), fmt.Sprintf(format, args))
	}
}

// Write the given line, prefixed by the given source.
// u may be nil, in which case, the line will be from the server.
// A line ending will be automatically appended.
func (c *Client) WriteFrom(u *core.User, format string, args ...interface{}) {
	if u != nil {
		fmt.Fprintf(c, ":%s!%s@%s %s\r\n", u.Nick(),
		            u.GetIdent(), u.GetHostname(),
		            fmt.Sprintf(format, args))
	} else {
		fmt.Fprintf(c, ":%s %s\r\n", "Server.name",
		            fmt.Sprintf(format, args))
	}
}
