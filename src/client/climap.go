package client

import "sync"

import "oddcomm/src/core"


// Counts our clients.
// We impose the terrible upper limit of four billion simultaneous clients.
var clicount uint32
var cliMutex sync.Mutex


// GetClient looks up a Client corresponding to a given User.
// If no such Client exists, or the Client is disconnecting, returns nil.
func GetClient(u *core.User) (c *Client) {

	// Check whether they're marked as ours before getting their struct.
	if u.Owner() != "oddcomm/src/client" {
		return nil
	}

	return u.Owndata().(*Client)
}

// Add a client to the client map.
func addClient(c *Client) {
	cliMutex.Lock()
	clicount++
	cliMutex.Unlock()
}

// Delete a client from the client map.
func delClient(c *Client) {
	cliMutex.Lock()
	clicount--
	
	// Poke the client subsystem goroutine if this is the last one, so it
	// knows that shutdown is okay, if it wants to shut down.
	if clicount == 0 {
		subsysMsg <- "clients gone"
	}

	cliMutex.Unlock()
}
