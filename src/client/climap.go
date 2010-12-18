package client

import "sync"

import "oddcomm/src/core"


// Contains the current client objects. When empty, we don't have any anymore.
// True is the normal value for a client. False means they're disconnecting.
var climap map[*Client]bool
var clients_by_user map[*core.User]*Client
var cliMutex sync.Mutex


func init() {
	climap = make(map[*Client]bool)
	clients_by_user = make(map[*core.User]*Client)
}


// GetClient looks up a Client corresponding to a given User.
// If no such Client exists, or the Client is disconnecting, returns nil.
func GetClient(u *core.User) (c *Client) {
	cliMutex.Lock()
	c = clients_by_user[u]
	cliMutex.Unlock()
	return
}

// Add a client to the client map.
func addClient(c *Client) {
	cliMutex.Lock()
	climap[c] = true
	clients_by_user[c.u] = c
	cliMutex.Unlock()
}

// Mark a client as disconnecting.
// Only to be called holding the client's Mutex.
func killClient(c *Client) {
	cliMutex.Lock()
	climap[c] = false
	clients_by_user[c.u] = nil, false
	cliMutex.Unlock()
}
// Delete a client from the client map.
func delClient(c *Client) {
	cliMutex.Lock()
	climap[c] = false, false
	cliMutex.Unlock()

	// Poke the client subsystem goroutine so it can shut us down if this
	// was the last one.
	subsysMsg <- "client deleted"
}
