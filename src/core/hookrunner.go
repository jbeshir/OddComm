package core


// Generic hook runner function.
var hookRunner chan func()


func init() {
	// Create simple worker goroutine.
	// The buffer here is the maximum length, recursively, of hooks invoked
	// by the same event. If it overflows, then the server will hang.
	hookRunner = make(chan func(), 1000)
	go runhooks()
}

func runhooks() {
	// Constantly read functions to run and run them.
	for {
		(<-hookRunner)()
	}
}
