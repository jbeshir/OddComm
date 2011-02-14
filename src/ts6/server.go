package ts6

import "fmt"
import "net"
import "os"
import "sync"

import "oddcomm/src/core"


// The server struct contains information on a server.
type server struct {
	sid  string  // The server's unique ID.
	name string  // The server's name.
	desc string  // The server's description.
	bursted bool // Whether it has finished sending its burst.
	up   *server // Parent server.
	down *server // First child server.
	prev *server // Previous sibling server.
	next *server // Next sibling server.
	local *local // The local server that introduced this server.
	mutex  sync.Mutex
}

// The local struct contains the state for directly linked servers.
type local struct {
	server                     // Embed information on this server.
	c             *net.TCPConn // This server's connection.
	authed        bool         // Whether this server has authenticated to us.
	auth_sent     bool         // Whether we've authenticated to it.
	burst_sent    bool         // Whether we've finished sending our burst.
	pass          string       // The server's sent password, for password auth.
	disconnecting bool         // Whether the server is currently disconnecting.
}

// Delete the server, with the given message.
// Repeated calls do nothing.
func (l *local) Delete(msg string) {
	l.mutex.Lock()
	l.delete(msg)
	l.mutex.Unlock()
}

// Delete the server.
// Assumes it is being called with the server mutex already held.
// Repeated calls do nothing.
func (l *local) delete(msg string) {
	serverMutex.Lock()

	if !l.disconnecting {
		fmt.Fprintf(l.c, "ERROR :%s\r\n", msg)
		l.c.Close()
		l.disconnecting = true

		if l.sid != "" {
			core.ReleaseSID(l.sid)
		}

		if l.prev != nil {
			l.prev.next = l.next
		} else {
			servers = l.next
		}

		if l.next != nil {
			l.next.prev = l.prev
		}
	}

	serverMutex.Unlock()
}

// Write to this local server.
// Is blocking, but threadsafe.
func (l *local) Write(b []byte) (n int, err os.Error) {
	l.mutex.Lock()

	if l.disconnecting {
		n, err = 0, os.NewError("Server disconnecting.")
	} else {
		n, err = l.c.Write(b)
		if err != nil {
			l.delete(err.String())
		}
	}

	l.mutex.Unlock()
	return
}

// Send a formatted line from the given user or server,
// to the given user or server.
func (l *local) SendLine(source, target interface{}, cmd string, format string, args ...interface{}) {
	fmt.Fprintf(l, ":%s %s %s %s\r\n", prefix(source), cmd, prefix(target),
		fmt.Sprintf(format, args...))
}

// Write a given prewritten line from the given user or server.
// source may be a nil interface or a nil value, in which case the line will
func (l *local) SendFrom(source interface{}, format string, args ...interface{}) {
	fmt.Fprintf(l, ":%s %s\r\n", prefix(source),
		fmt.Sprintf(format, args...))
}


// Returns the prefix to be used for a line from the given user or server.
// This may be a nil interface or nil value, in which case the line will be
// from this server.
func prefix(source interface{}) string {
	if u, ok := source.(*core.User); ok && u != nil {
		return u.ID()
	} else if s, ok := source.(*server); ok && s != nil {
		return s.sid
	} else if s, ok := source.(string); ok && s != "" {
		return s
	}
	return "1AA"
}
