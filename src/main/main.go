/*
	The main package for the server.

	Starts every package goroutine, then waits for them to shut down
	before terminating.
*/
package main

import "oddcomm/src/core"
import "oddcomm/lib/persist"

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
	var exitList []chan int
	var msg chan string
	var exit chan int

	// Load saved state and configuration.
	// STUB: This should try to open a saved state file.
	persist.FirstRun()

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

	// Save state and configuration.
	// STUB: This should try to open and save to a file.
}
