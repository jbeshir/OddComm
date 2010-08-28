/*
	Dummy mode-adding module, to demonstrate the utter triviality of it.
	Also tests non-ASCII modes, demonstrating how they break clients.
	(They do. Don't actually USE this module.)
*/
package tmmode

import "oddcomm/src/client"


// Must be set, must be unique.
var MODULENAME string = "dev/tmmode"


func init() {
	// Add the mode.
	client.ChanModes.AddSimple('™', "trademarked")
}
