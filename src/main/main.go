/*
	The main package for the server.

	Starts every package goroutine, then waits for them to shut down
	before terminating.
*/
package main

import "oddircd/src/core"
import "oddircd/src/client"

import modules_client_botmode "oddircd/modules/client/botmode"
import modules_client_extbans "oddircd/modules/client/extbans"
import modules_dev_catserv "oddircd/modules/dev/catserv"
import modules_dev_horde "oddircd/modules/dev/horde"
import modules_dev_tmmode "oddircd/modules/dev/tmmode"

func main() {
	// Makes modules be permitted to link in.
	_ = modules_client_botmode.MODULENAME
	_ = modules_client_extbans.MODULENAME
	_ = modules_dev_catserv.MODULENAME
	_ = modules_dev_horde.MODULENAME
	_ = modules_dev_tmmode.MODULENAME

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
