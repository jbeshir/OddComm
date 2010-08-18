package client

import "oddircd/src/core"
import "oddircd/src/irc"


// UserModes is the mode parser for user modes.
// Modes can be added to this during init() only.
// It is safe in modules to overwrite default modes with your own, but modes
// extended in core must be extended in your module.
var UserModes *irc.ModeParser

// ChanModes is the mode parser for channel modes.
// Modes can be added to this during init() only.
// It is safe in modules to overwrite default modes with your own, but modes
// extended in core must be extended in your module.
var ChanModes *irc.ModeParser


func init() {
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

	// Add channel modes +i, +m, and +s.
	ChanModes.AddSimple('i', "invite-only")
	ChanModes.AddSimple('m', "muted")
	ChanModes.AddSimple('s', "hidden")

	// Add ban, ban exception, and unrestrict (invex) modes on channels.
	ChanModes.AddList('b', "ban host")
	ChanModes.AddList('e', "banexception host")
	ChanModes.AddList('I', "unrestrict host")

	// DELIBERATELY NOT IMPLEMENTED: +k, +p
	// These modes are not deemed to be the optimal way of doing anything.
}
