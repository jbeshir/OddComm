/*
	The main package for the server.

	Starts every package goroutine, then waits for them to shut down
	before terminating.
*/
package main

import "oddircd/src/core"
import "oddircd/src/client"


func main() {
	var exitList [1]chan int
	var msg chan string
	var exit chan int

	// Start client subsystem.
	msg, exit = client.Start()
	if msg != nil {
		core.AddPackage("oddircd/src/client", msg)
	}
	exitList[0] = exit

	// Run start hooks.
	core.RunStartHooks()

	// Wait until every package goroutine returns.
	for i := range exitList {
		if exitList[i] != nil {
			<-exitList[i]
		}
	}
}
