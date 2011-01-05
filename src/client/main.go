/*
	Implements introduction of and communication with users connecting
	with an IRC client.

	Uses an RFC 1459-derived protocol with compatibility with the majority
	of popular clients being the primary goal.
*/
package client

import "fmt"
import "net"
import "sync"

import "oddcomm/src/core"

// Create a channel for sending messages to the subsystem's goroutine.
var subsysMsg chan string = make(chan string)

// Counts our clients.
// We impose the terrible upper limit of four billion simultaneous clients.
var clicount uint32
var exiting bool
var cliMutex sync.Mutex


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

			cliMutex.Lock()

			// Note that we're terminating, as soon as
			// every client is done quitting.
			exiting = true

			// Stop every current client by deleting their user.
			core.IterateUsers("oddcomm/src/client",
				func(u *core.User) {
					u.Delete(u, "Server Terminating")
				})
		}

		// If we're exiting, see if we've no more clients. If we've no
		// more clients, we're good to shut down as soon as everyone
		// else is, so send the exit signal and wait for the end.
		if exiting {
			cliMutex.Lock()
			if clicount == 0 {
				exit <- 0
				exiting = false
			}
			cliMutex.Unlock()
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

		client.u = core.NewUser("oddcomm/src/client", client, false, "", data)
		incClient(client)

		go input(client)
	}
}


// Increment client count.
// If we're exiting, kills the client immediately.
// Must be called while holding the client mutex, and after adding the client's
// user to the server, so on shutdown the client will guaranteeably either have
// been already added to the user list and killed there, or will be killed here.
func incClient(c *Client) {
	cliMutex.Lock()
	if !exiting {
		clicount++
	} else {
		c.delete("Server Terminating")
	}
	cliMutex.Unlock()
}

// Decrement client count.
func decClient(c *Client) {
	cliMutex.Lock()
	clicount--
	
	// Poke the client subsystem goroutine if this is the last one, so it
	// knows that shutdown is okay, if it wants to shut down.
	if clicount == 0 {
		subsysMsg <- "clients gone"
	}

	cliMutex.Unlock()
}


// GetClient looks up a Client corresponding to a given User.
// If no such Client exists, or the Client is disconnecting, returns nil.
func GetClient(u *core.User) (c *Client) {

	// Check whether they're marked as ours before getting their struct.
	if u.Owner() != "oddcomm/src/client" {
		return nil
	}

	return u.Owndata().(*Client)
}
