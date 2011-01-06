package client

import "oddcomm/lib/irc"

// Commands is the client command dispatcher.
// The source for any command handlers called here will be a Client pointer.
// Commands can be added to this during init() only, to add commands to the
// client subsystem.
var Commands irc.CommandDispatcher
