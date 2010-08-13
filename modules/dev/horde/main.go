/*
	Add a massive horde (100k) of test users.
*/
package catserv

import "oddircd/src/core"


// Must be set, must be unique.
var MODULENAME string = "dev/horde"


func init() {
	// Join the server on startup.
	core.HookStart(addHorde)
}

func addHorde() {
	// Add the horde.
	for i := 0; i < 100000; i++ {
		core.NewUser("oddircd/modules/dev/horde", true, "")
	}
}
