/*
	Implements introduction of and communication with users connecting
	with an IRC client.

	Uses an RFC 1459-derived protocol with compatibility with the majority
	of popular clients being the primary goal.
*/
package client

import "fmt"
import "net"

import "oddcomm/src/core"

// Create a channel for sending messages to the subsystem's goroutine.
var subsysMsg chan string = make(chan string)


// Start starts up the client subsystem.
func Start() (msg chan string, exit chan int) {
	msg = subsysMsg
	exit = make(chan int)

	go clientMain(msg, exit)

	return
}


// Main function for the package.
// Starts up the listening goroutines, performs message handling.
// Owns the client map.
func clientMain(msg chan string, exit chan int) {

	// Generate the version string.
	for hook := supportHooks; hook != nil; hook = hook.next {
		supportLine += hook.h()
	}

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
		// Handle messages to the goroutine.
		message := <-msg

		// If asked to exit...
		if message == "exit" {

			// Stop the listening goroutine.
			if l != nil {
				l.Close()
			}

			// Stop every client.
			for c := range climap {
				c.mutex.Lock()
				c.delete("Server Terminating")
				c.mutex.Unlock()
			}

			// Note that we're terminating, as soon as
			// every client is done quitting.
			exiting = true
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
func listen(l *net.TCPListener) {

	for {
		c, err := l.AcceptTCP()
		if err != nil {
			core.Shutdown()
			return
		}

		client := new(Client)
		client.conn = c
		client.conn.SetWriteTimeout(1000)

		ip := client.conn.RemoteAddr().(*net.TCPAddr).IP.String()
		data := make([]core.DataChange, 2)
		data[0].Name, data[0].Data = "ip", ip
		data[1].Name, data[1].Data = "hostname", ip
		data[0].Next = &data[1]
		
		client.u = core.NewUser("oddcomm/src/client", false, "", &data[0])

		addClient(client)

		go input(client)
	}
}
