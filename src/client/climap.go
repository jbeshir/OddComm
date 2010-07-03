package client

// Contains the current client objects. When empty, we don't have any anymore.
// True is the normal value for a client. False means they're disconnecting.
var climap map[*Client]bool


func init() {
	climap = make(map[*Client]bool)
}


// Add a client to the client map.
func addClient(c *Client) {
	makeRequest(nil, func() {
		climap[c] = true
	})
}

// Mark a client as disconnecting.
// Only to be called by the client's own goroutine.
func killClient(c *Client) {
	makeRequest(nil, func() {
		climap[c] = false
	})
}
// Delete a client from the client map.
func delClient(c *Client) {
	makeRequest(nil, func() {
		climap[c] = false, false
	})
}
