/*
	The TS6 package provides traditional IRC linking using the TS6 protocol.
*/
package ts6

import "fmt"
import "net"
import "sync"

import "oddcomm/lib/irc"
import "oddcomm/src/core"


var me string = "ts6"


// An ordering is defined on mutexes in this package; only one server mutex may be held
// at once, and if the server list mutex is also held, it must be locked after the
// specific server mutex. This prevents deadlocks.


// Points to the first server.
var servers *server
var serverMutex sync.Mutex

// Commands added here will be called with either a server or a core.User.
var commands = irc.NewCommandDispatcher()

// Create a channel for sending messages to the subsystem's goroutine.
var subsysMsg chan string = make(chan string)


// Start starts up the TS6 subsystem.
func Start() (msg chan string, exit chan int) {
	msg = subsysMsg
	exit = make(chan int)

	go ts6_main(msg, exit)

	return
}

func ts6_main(msg chan string, exit chan int) {

	// No configuration, so this is fixed.
	var addr net.TCPAddr
	addr.IP = net.IPv4(127, 0, 0, 1)
	addr.Port = 3725

	// Start our listener.
	l, err := net.ListenTCP("tcp4", &addr)
	if err != nil {
		fmt.Printf("No bind: %s\n", err)
		exit <- 0
	} else {
		go listen(l)

		// Again, no configuration.
		addr.Port = 13725
		c, err := net.DialTCP("tcp4", nil, &addr)
		if err != nil {
			fmt.Printf("No connection: %s\n", err)
		} else {
			go link(c, true)
		}
	}

	// Handle messages.
	for message := range msg {
		if message == "exit" {
			exit <- 0
		}
	}
}

// Listen goroutine function, handling listening for one socket.
// Owns its socket.
func listen(l *net.TCPListener) {
	for {
		c, err := l.AcceptTCP()
		if err != nil {
			core.Shutdown()
			return
		}
		go link(c, false)
	}
}
