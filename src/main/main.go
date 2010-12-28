/*
	The main package for the server.

	Starts every package goroutine, then waits for them to shut down
	before terminating.
*/
package main

import "oddcomm/src/core"
import "oddcomm/src/client"
import "oddcomm/src/ts6"

import _     "oddcomm/modules/user/botmark"
import _   "oddcomm/modules/client/extbans"
import _     "oddcomm/modules/client/login"
import _ "oddcomm/modules/client/ochanctrl"
import _  "oddcomm/modules/client/opermode"
import _   "oddcomm/modules/client/opflags"
import _     "oddcomm/modules/oper/account"
import _  "oddcomm/modules/oper/pmoverride"
import _      "oddcomm/modules/dev/catserv"
import _        "oddcomm/modules/dev/horde"
import _  "oddcomm/modules/dev/testaccount"
import _       "oddcomm/modules/dev/tmmode"

func main() {
	exitList := make([]chan int, 0)
	var msg chan string
	var exit chan int

	// Start client subsystem.
	msg, exit = client.Start()
	if msg != nil {
		core.AddPackage("oddcomm/src/client", msg)
	}
	exitList = append(exitList, exit)

	// Start TS6 subsystem.
	msg, exit = ts6.Start()
	if msg != nil {
		core.AddPackage("oddcomm/src/ts6", msg)
	}
	exitList = append(exitList, exit)

	// Run start hooks.
	core.RunStartHooks()

	// Wait until every package goroutine returns.
	for _, exit := range exitList {
		if exit != nil {
			<-exit
		}
	}
}
