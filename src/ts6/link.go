package ts6

import "net"


// Handle a single (potential) server link.
// outgoing indicates whether it is outgoing or incoming.
func link(c *net.TCPConn, outgoing bool) {
	c.Close()
}
