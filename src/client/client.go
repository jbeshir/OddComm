package client

import "fmt"
import "net"
import "os"

import "oddircd/core"


// Handle a client connection.
type Client struct {
	cchan     chan clientRequest
	conn      *net.TCPConn
	u         *core.User
	disconnecting bool
	inputDone bool
	outbuf    []byte
}

func (c *Client) Remove(message string) {
	makeRequest(c, func() {
		c.remove(message)
	})
}

func (c *Client) remove(message string) {
	if c.u != nil {
		c.u.Remove(message)
	}

	username := c.u.Data("ircident")
	if username == "" {
		username = "unknown"
	}
	c.write([]byte(fmt.Sprintf("ERROR :Closing link: (%s@%s) [%s]\r\n", username, c.u.Data("hostname"), message)))

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
			if (c.u != nil) {
				c.u.Remove("SendQ exceeded.")
			}
			c.disconnecting = true
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
			c.disconnecting = true
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
// Privmsg sends a PRIVMSG to the client.
func (c *Client) Privmsg(source string, message []byte) {
	fmt.Fprintf(c, "%s PRIVMSG Namegduf :%s", source, message)
}
