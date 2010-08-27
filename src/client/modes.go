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
	UserModes = irc.NewModeParser(false)
	ChanModes = irc.NewModeParser(false)

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

	// Add user mode +o.
	UserModes.AddSimple('o', "op")

	// Add channel modes +i, +m, and +s.
	ChanModes.AddSimple('i', "restrict join")
	ChanModes.AddSimple('m', "restrict mute")
	ChanModes.AddSimple('s', "hidden")

	// Add channel ban, ban exception, and unrestrict (invex) modes.
	ChanModes.AddList('b', "ban host")
	ChanModes.AddList('e', "banexception host")
	ChanModes.AddList('I', "unrestrict host")

	// Add channel op and voice membership modes.
	ChanModes.AddMembership('o', "op")
	ChanModes.AddMembership('v', "voiced")

	// DELIBERATELY NOT IMPLEMENTED
	// User modes: +i (always on)
	// Channel modes: +k, +l, +n (always on), +p
	// These modes are not deemed to be the optimal way of doing anything.
}
