package ts6

import "fmt"
import "net"
import "os"
import "sync"

import "oddcomm/src/core"


// The server struct contains information on a server.
type server struct {
	u    *core.User
	up   *server // Parent server.
	down *server // First child server.
	next *server // Next sibling server.
	local *local // Local server in this server's direction.
	mutex  sync.Mutex
}

// The local struct contains the state for directly linked servers.
type local struct {
	s      server
	c      *net.TCPConn
	authed bool
	pass   string
}


// Write to this local server.
func (l *local) Write(b []byte) (n int, err os.Error) {
	l.s.mutex.Lock()
	n, err = l.c.Write(b)
	l.s.mutex.Unlock()
	return
}

// Write a formatted line from the given user, addressed to this server.
// u may be nil, in which case, the line will be from this server.
// A line ending will be automatically appended.
func (l *local) WriteTo(u *core.User, cmd string, format string, args ...interface{}) {
	if u != nil {
		fmt.Fprintf(l, ":%s %s %s %s\r\n", u.ID(), cmd,
			l.s.u.ID(), fmt.Sprintf(format, args...))
	} else {
		fmt.Fprintf(l, ":%s %s %s %s\r\n", "0ZZ", cmd,
			l.s.u.ID(), fmt.Sprintf(format, args...))
	}
}

// Write the given line, prefixed by the given source.
// u may be nil, in which case, the line will be from this server.
// A line ending will be automatically appended.
func (l *local) WriteFrom(u *core.User, format string, args ...interface{}) {
	if u != nil {
		fmt.Fprintf(l, ":%s %s\r\n", u.ID(),
			fmt.Sprintf(format, args...))
	} else {
		fmt.Fprintf(l, ":%s %s\r\n", "0ZZ",
			fmt.Sprintf(format, args...))
		}
}
