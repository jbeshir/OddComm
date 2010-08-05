package client

import "oddircd/irc"


// The client command dispatcher.
// Commands can be added to this during init() only, to add commands to the
// client subsystem.
var Commands irc.CommandDispatcher

func init() {
	Commands = irc.NewCommandDispatcher()
}
