package logic

import "net"

import "oddcomm/src/core/connect"
import "oddcomm/src/core/connect/mmn"

// List of all nodes.
var Nodes []*Node

// Our own Node ID.
var Id uint16

// Our own Node.
var Me *Node

// Represents a node.
type Node struct {
	*connect.ConnInfo      // Connection information for the node.
	Id      uint16         // Node ID.
	NewConn chan net.Conn  // Channel to send incoming connections to.
	conn    *connect.Conn  // Current connection. Nil if none.
	queue   []*mmn.Line    // Line queue.
	receive chan *mmn.Line // Channel received lines are sent to.
	send    chan *mmn.Line // CHannel lines to be sent are sent to.
	waiting net.Conn       // Connection waiting for prev conn to die.
}

// Create a new node with the given ID and address.
// Our own node ID must be set before creating nodes.
func NewNode(id uint16, connInfo *connect.ConnInfo) *Node {

	n := new(Node)
	n.ConnInfo = connInfo
	n.Id = id

	n.send = make(chan *mmn.Line, 10)
	n.NewConn = make(chan net.Conn, 10)

	// Add to node list.
	Nodes = append(Nodes, n)

	// If this node is ourselves, set it as ours.
	if n.Id == Id {
		Me = n
		n.receive = make(chan *mmn.Line, 10)
	}

	// Start processing lines to/from this node.
	go n.process()

	return n
}

// The node's goroutine. Handle lines sent to or received from this node.
// TODO: Figure out how not to deadlock when two nodes send to each other.
// Use an intermediary?
func (n *Node) process() {
	for {
		select {

		// Handle a received line on our connection.
		case line, ok := <-n.receive:

			if ok {
				// Process the line.
				// STUB
				line = line
				continue
			}

			// Connection closed.
			n.receive = nil // Stop us selecting on closed chan.

			// If we have a waiting incoming connection, take that.
			if n.waiting != nil {
				n.conn = connect.NewIncoming(n.waiting)
				n.waiting = nil
				n.receive = make(chan *mmn.Line, 10)
				go n.conn.ReadLines(n.receive)

				// SEND INITIAL CONNECTION MESSAGE

				continue
			}

			// Otherwise, try to make a new connection.
			var err error
			n.conn, err = connect.NewOutgoing(n.ConnInfo)
			if err == nil {
				n.receive = make(chan *mmn.Line, 10)
				go n.conn.ReadLines(n.receive)
			}

		// Handle a line to be sent on our connection.
		case line := <-n.send:

			// Send the line.
			n.sendSyncLine(line)

		// Handle a new connection from this node.
		case conn := <-n.NewConn:

			// If we have no existing connection, adopt this one.
			if n.conn == nil {
				n.conn = connect.NewIncoming(conn)
				n.receive = make(chan *mmn.Line, 10)
				go n.conn.ReadLines(n.receive)

				// SEND INITIAL CONNECTION MESSAGE

				continue
			}

			// If our connection's state is not a new outgoing,
			// close it, wait for reading to end, then adopt
			// this as our connection.
			if n.conn.State != connect.ConnStateInitialOutgoing {
				n.conn.Close()
				n.waiting = conn
				continue
			}

			// TODO: What to do if it is?
			// Two outgoing connections crossing.
			// For now, close it; two connections exactly at once
			// can fail, but hopefully a retry will deal with it.
			conn.Close()
		}
	}
}

// Send a line to the node, once we have a synchronized connection to it.
// If the node is unreachable, we drop the line.
// Must be run from the node's goroutine.
func (n *Node) sendSyncLine(line *mmn.Line) {

	// If this node is ourselves, send it directly to our receive chan.
	if n == Me {
		n.receive <- line
		return
	}

	// If we have a synchronized connection now, send it now.
	if n.conn != nil && n.conn.State == connect.ConnStateNormal {
		err := n.conn.WriteLine(line)
		if err != nil {
			n.conn.Close()
		}
		return
	}

	// Otherwise, add it to the queue.
	n.queue = append(n.queue, line)

	// If we have no connection, make an attempt to establish one.
	// Drop the line if we get an error.
	if n.conn == nil {
		var err error
		n.conn, err = connect.NewOutgoing(n.ConnInfo)
		if err != nil {
			n.queue = n.queue[:0]
		}
	}
}


// Attempt an outgoing connection to all nodes.
// Nodes which already have a connection are skipped.
// Our nodes must all be added before this is called.
func StartOutgoing() {
	for _, n := range Nodes {

		// If this is ourselves, skip.
		if n == Me {
			continue
		}

		// If they already have a connection, skip.
		if n.conn != nil {
			continue
		}

		// Otherwise, attempt an outgoing connection.
		var err error
		n.conn, err = connect.NewOutgoing(n.ConnInfo)
		if err == nil {
			n.receive = make(chan *mmn.Line, 10)
			go n.conn.ReadLines(n.receive)
		}
	}
}
