package client


type clientRequest struct {
	c    *Client
	f    func()
	done chan bool
}

// Requests that f be run from the goroutine belonging to the c.
// If c == nil, requests that f be run from the client main goroutine instead.
// f will not be called if c is disconnecting or has already disconnected.
// This function waits until the function has either been called, or discarded,
// and returns whether it was called successfully or not.
func makeRequest(c *Client, f func()) (run bool) {
	var r clientRequest
	r.c = c
	r.f = f
	r.done = make(chan bool)

	clichan <- r
	run = <-r.done

	return
}


// Sends a direct request to the goroutine owning c to to run f.
// Said goroutine must be known to be alive, making this only safe to use from
// the same client's input or output goroutine.
// This function waits until f has been called before returning.
func makeDirectRequest(c *Client, f func()) {
	var r clientRequest
	r.f = f
	r.done = make(chan bool)

	c.cchan <- r
	<-r.done
}
