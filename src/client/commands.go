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

// ChanModes is the mode parser for channel modes.
// Modes can be added to this during init() only.
var ChanModes *irc.ModeParser

func init() {
	Commands = irc.NewCommandDispatcher()
	UserModes = irc.NewModeParser()
	ChanModes = irc.NewModeParser()

	// Fake an always-on +i on users.
	UserModes.AddSimple('i', "__placeholder__")
	UserModes.AddExtMode('i', "", func(_ bool, _ core.Extensible,
	                                   _ string) *core.DataChange {
		return nil
	}, nil, func(_ core.Extensible) string {
		return "on"
	})

	// Fake an always-on +n on channels.
	ChanModes.AddSimple('n', "__placeholder__")
	ChanModes.AddExtMode('n', "", func(_ bool, _ core.Extensible,
	                                    _ string) *core.DataChange {
		return nil
	}, nil, func(_ core.Extensible) string {
		return "on"
	})
}
