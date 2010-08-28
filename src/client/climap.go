package client

import "oddcomm/src/core"


// Contains the current client objects. When empty, we don't have any anymore.
// True is the normal value for a client. False means they're disconnecting.
var climap map[*Client]bool
var clients_by_user map[*core.User]*Client


func init() {
	climap = make(map[*Client]bool)
	clients_by_user = make(map[*core.User]*Client)
}


// GetClient looks up a Client corresponding to a given User.
// If no such Client exists, or the Client is disconnecting, returns nil.
func GetClient(u *core.User) (c *Client) {
	makeRequest(nil, func() {
		c = clients_by_user[u]
	})
	return
}

// Add a client to the client map.
func addClient(c *Client) {
	makeRequest(nil, func() {
		climap[c] = true
		clients_by_user[c.u] = c
	})
}

// Mark a client as disconnecting.
// Only to be called by the client's own goroutine.
func killClient(c *Client) {
	makeRequest(nil, func() {
		climap[c] = false
		clients_by_user[c.u] = nil, false
	})
}
// Delete a client from the client map.
func delClient(c *Client) {
	makeRequest(nil, func() {
		climap[c] = false, false
	})
}
