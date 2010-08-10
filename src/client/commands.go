package client

import "oddircd/src/core"
import "oddircd/src/irc"


// Commands is the client command dispatcher.
// Commands can be added to this during init() only, to add commands to the
// client subsystem.
var Commands irc.CommandDispatcher

// UserModes is the mode parser for user modes.
// Modes can be added to this during init() only.
var UserModes *irc.ModeParser

func init() {
	Commands = irc.NewCommandDispatcher()
	UserModes = irc.NewModeParser()

	UserModes.AddSimple('i', "__placeholder__")
	UserModes.AddExtMode('i', "", func(_ bool, _ core.Extensible,
	                                   _ string) *core.DataChange {
		return nil
	}, nil)
}
