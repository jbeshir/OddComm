/*
	The main package for the server.

	Starts every package goroutine, then waits for them to shut down
	before terminating.
*/
package main

import "oddircd/src/core"
import "oddircd/src/client"

import modules_catserv "oddircd/modules/catserv"
import modules_irc_botmode "oddircd/modules/irc/botmode"

func main() {
	// Makes modules be permitted to link in.
	_ = modules_catserv.MODULENAME
	_ = modules_irc_botmode.MODULENAME

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
