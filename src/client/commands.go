package client

import "oddcomm/lib/irc"

// Commands is the client command dispatcher.
// Commands can be added to this during init() only, to add commands to the
// client subsystem.
var Commands irc.CommandDispatcher


func init() {
	Commands = irc.NewCommandDispatcher()
}
