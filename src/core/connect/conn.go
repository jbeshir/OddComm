package connect

import "net"
import "sync"
import "crypto/tls"

import proto "goprotobuf.googlecode.com/hg/proto"

import "oddcomm/src/core/connect/mmn"

var links = make(map[uint16]*Conn)
var linksMutex sync.Mutex


// Represents an arbitrary stream which mmn.Lines can be written to.
type LineStream interface {

	// Writes the given line to this stream.
	WriteLine(line mmn.Line) error

	// Closes the stream. Does not return an error;
	// using code must tolerate unclean termination without any special handling.
	Close()
}


// Represents a connection to a node.
type Conn struct {
	net.Conn
	Id uint16
	Mutex sync.Mutex
	Synchronised bool
}

// Gets a connection to the given node.
//
// Makes a new connection if none already exists,
// returns an existing connection otherwise.
//
// Returns the connection, and whether it was newly created or not.
func GetConn(id uint16, addr string) (*Conn, bool, error) {

	linksMutex.Lock()

	c, ok := links[id]

	var err error
	if !ok {
		c = new(Conn)
		c.Id = id

		c.Conn, err = tls.Dial("tcp", addr, nil)
		if err == nil {
			links[id] = c
		}
	}

	linksMutex.Unlock()

	if ok {
		return c, true, nil
	} else if err == nil {
		return c, false, nil
	}

	return nil, false, err
}

// Write an mmn.Line to the connection.
func (c *Conn) WriteLine(line mmn.Line) (err error) {

	c.Mutex.Lock()

	var buf []byte
	buf, err = proto.Marshal(line)
	if (err != nil) {
		panic("Error marshalling protobuf struct.")
	}

	for (len(buf) > 0) {
		var n int
		n, err = c.Conn.Write(buf)
		buf = buf[n:]

		if (err != nil) {
			c.Close()
			break
		}
	}

	c.Mutex.Unlock()
	return
}

// Close the connection.
func (c *Conn) Close() {

	c.Mutex.Lock()

	c.Conn.Close()

	linksMutex.Lock()
	delete(links, c.Id)
	linksMutex.Unlock()

	c.Mutex.Unlock()
}
