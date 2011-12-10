package connect

import "net"
import "sync"
import "crypto/tls"
import "crypto/x509"

import proto "goprotobuf.googlecode.com/hg/proto"

import "oddcomm/src/core/connect/mmn"

// Our tls certificate for authentication.
// Must be set before connections are made.
var Cert tls.Certificate

// Represents a current state of a connection to another node.
type ConnState int

const (
	// Version negotiation.
	// Incoming means we've sent a list and are waiting for version.
	// Outgoing means we're waiting to receive the list to send version.
	ConnStateInitialIncoming ConnState = iota
	ConnStateInitialOutgoing ConnState = iota

	// Incoming means we've sent cap and are waiting to receive one.
	// Outgoing means we're waiting to receive before sending. 
	ConnStateCapabilityNegotiationIncoming ConnState = iota
	ConnStateCapabilityNegotiationOutgoing ConnState = iota

	// Waiting to receive a degraded notification from the other end.
	ConnStateDegradedNegotiation ConnState = iota

	// Waiting to receive a nonce for synchronization.
	ConnStateSynchronization ConnState = iota

	// We're presently receiving a burst.
	ConnStateReceivingBurst ConnState = iota

	// We're presently waiting to send or sending a burst.
	// We expect to receive nothing but ping/pong while doing this.
	ConnStateWaitingToSendBurst ConnState = iota
	ConnStateSendingBurst       ConnState = iota

	// Normal operating state.
	ConnStateNormal ConnState = iota

	// Connection has been closed.
	ConnStateClosed ConnState = iota
)


// Represents a connection to a node.
type Conn struct {
	Id           uint16
	Synchronised bool
	State        ConnState

	conn  net.Conn
	mutex sync.Mutex
}

// Creates a new outgoing connection to the given address.
func NewOutgoing(info *ConnInfo) (*Conn, error) {

	tlsConfig := new(tls.Config)
	tlsConfig.Certificates = append([]tls.Certificate(nil), Cert)
	tlsConfig.RootCAs = info.Cert

	tlsConn, err := tls.Dial("tcp", info.Addr, tlsConfig)
	if err != nil {
		return nil, err
	}

	c := new(Conn)
	c.conn = tlsConn
	c.State = ConnStateInitialOutgoing

	return c, nil
}

// Creates a new incoming connection.
func NewIncoming(conn net.Conn) *Conn {

	c := new(Conn)
	c.conn = conn
	c.State = ConnStateInitialIncoming

	return c
}

// Write an mmn.Line to the connection.
func (c *Conn) WriteLine(line *mmn.Line) (err error) {

	var buf []byte
	buf, err = proto.Marshal(line)
	if err != nil {
		panic("Error marshalling protobuf struct.")
	}

	for len(buf) > 0 {
		var n int
		n, err = c.conn.Write(buf)
		buf = buf[n:]

		if err != nil {
			c.Close()
			break
		}
	}

	return
}

// Close the connection.
func (c *Conn) Close() {

	c.conn.Close()
	c.State = ConnStateClosed
}


// Represents connection information for a node.
type ConnInfo struct {
	Addr string
	Cert *x509.CertPool
}
