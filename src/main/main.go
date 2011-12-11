/*
	The main package for the server.

	Starts every package goroutine, then waits for them to shut down
	before terminating.
*/
package main

import "flag"

import "oddcomm/src/core"
//import "oddcomm/lib/persist"
/*
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
*/

var stateFile = "oddcomm.state"

func main() {

	// Define and parse flags.
	id := flag.Uint("id", 0, "Set the node ID of this OddComm instance.")
	flag.Parse()

	// Validate flags.
	if (*id == 0 || *id > 0xFFFF) {
		panic("Invalid node id specified.")
	}

	// Start the core.
	core.Initialize(uint16(*id))

	/*
	var exitList []chan int
	var msg chan string
	var exit chan int

	// Load saved state and configuration.
	if f, err := os.Open(stateFile); err == nil {
		fmt.Printf("Loading previous settings...\n")
		if err := persist.Load(f); err != nil {
			fmt.Printf("Error loading settings: %s\n", err)
			return
		}
	} else {
		if v, ok := err.(*os.PathError); !ok || v.Error != os.ENOENT {
			fmt.Printf("Error loading state file: %s\n", err)
			fmt.Printf("Terminating.\n")
			return
		}
		fmt.Printf("Loading default settings...\n")
		persist.FirstRun()
	}

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
	// Load saved state and configuration.
	fmt.Printf("Saving settings...\n")
	var f *os.File
	var err os.Error
	if f, err = os.Create(stateFile + ".tmp"); err == nil {
		if err == nil {
			err = persist.Save(f)
			if err == nil {
				err = os.Rename(stateFile + ".tmp", stateFile)
			}
		}
	}
	if err != nil {
		fmt.Printf("Error saving settings: %s\n", err)
	} else {
		fmt.Printf("Settings saved. Terminating.\n")
	}
	*/

	// Wait forever.
	// TODO: Restore ability to terminate cleanly.
	select {}
}
