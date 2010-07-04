/*
	Implements introduction of and communication with users connecting
	with an IRC client.

	Uses an RFC 1459-derived protocol with compatibility with the majority
	of popular clients being the primary goal.
*/
package client

import "fmt"
import "net"

import "oddircd/core"


// Channel to send requests to run something from a client's goroutine to.
var clichan chan clientRequest


// Start starts up the client subsystem.
func Start() (msg chan string, exit chan int) {
	msg = make(chan string)
	exit = make(chan int)

	clichan = make(chan clientRequest)

	go clientMain(msg, exit)

	return
}


// Main function for the package.
// Starts up the listening goroutines, performs message handling.
// Owns the client map.
func clientMain(msg chan string, exit chan int) {

	// Right now, we just bind to a fixed IP/port.
	// No config parsing.
	var addr net.TCPAddr
	addr.IP = net.IPv4(127, 0, 0, 1)
	addr.Port = 6667

	// Start our listener.
	l, err := net.ListenTCP("tcp4", &addr)
	if err != nil {
		fmt.Printf("No bind: %s\n", err)
		core.Shutdown()
	} else {
		go listen(l)
	}

	var exiting bool
	for {
		select {

		// Handle messages to the goroutine.
		case message := <-msg:

			// If asked to exit...
			if message == "exit" {

				// Stop the listening goroutine.
				if l != nil {
					l.Close()
				}

				// Stop every client.
				var r clientRequest
				var c *Client
				r.f = func() {
					c.u = nil
					c.unreg = nil
					c.write([]byte("Server terminating.\r\n"))
				}
				r.done = make(chan bool)
				for c = range climap {
					r.c = c
					c.cchan <- r
					<-r.done
				}

				// Note that we're terminating, as soon as
				// every client is done quitting.
				exiting = true
			}

		// Handle requests to be sent to a client goroutine.
		// These are forwarded if the client exists and is alive,
		// and dropped otherwise.
		// If c == nil, it's a request to run the function here, so
		// do so.
		case r := <-clichan:
			if r.c != nil {
				alive, ok := climap[r.c]
				if ok && alive {
					r.c.cchan <- r
				} else {
					r.done <- false
				}
			} else {
				r.f()
				r.done <- true
			}
		}

		// If we're exiting, see if we've no more clients. If we've no
		// more clients, we're good to shut down as soon as everyone
		// else is, so send the exit signal and wait for the end.
		if exiting {
			done := true
			for _ = range climap {
				done = false
				break
			}

			if done {
				exit <- 0
				exiting = false
			}
		}
	}
}


// Listen goroutine function, handling listening for one socket.
// Owns its socket.
func listen(l net.Listener) {

	for {
		c, err := l.Accept()
		if err != nil {
			core.Shutdown()
			return
		}

		client := new(Client)
		client.cchan = make(chan clientRequest)
		client.conn = c
		client.conn.SetWriteTimeout(1)
		client.unreg = new(unregUser)

		go clientHandler(client)
	}
}
