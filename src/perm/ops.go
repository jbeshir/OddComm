package perm

import "strings"

import "oddircd/src/core"


// Stores the default powers for channel and server op flags.
var defChanOp []string
var defServerOp []string


func init() {
	// Add the core default channel op flags.
	AddChanDefOpFlag("ban")
	AddChanDefOpFlag("banview")
	AddChanDefOpFlag("invite")
	AddChanDefOpFlag("moderate")
	AddChanDefOpFlag("msg")
	AddChanDefOpFlag("topic")
	AddChanDefOpFlag("(de)op")
	AddChanDefOpFlag("(de)voice")

	// Add the core default server op flags.
	AddServerDefOpFlag("ban")
	AddServerDefOpFlag("broadcast")
	AddServerDefOpFlag("msg")
}


// AddChanDefOpFlag adds the given flag to the default channel op flag list.
// Does nothing if the flag already is in the default list, so can safely be
// used repeatedly by different modules using the flag.
// called repeatedly by different modules using the flag.
// Can only be called during init.
func AddChanDefOpFlag(flag string) {
	newDefault := flag
	for _, word := range defChanOp {
		if word == flag {
			return
		}
		newDefault += " " + word
	}
	defChanOp = strings.Fields(newDefault)
}

// AddServerDefOpFlag adds the given flag to the default server op flag list.
// Does nothing if the flag already is in the default list, so can safely be
// called repeatedly by different modules using the flag.
// Can only be called during init.
func AddServerDefOpFlag(flag string) {
	newDefault := flag
	for _, word := range defServerOp {
		if word == flag {
			return
		}
		newDefault += " " + word
	}
	defServerOp = strings.Fields(newDefault)
}

// HasOpFlag returns whether the user has the given op flag.
// This both checks for the presence of the flag, and, if the flag is default,
// the "on" keyword for default privileges. This function should not be used as
// a direct means of determining a user's ability to do something; instead,
// the appropriate permission check should be used, as this permits module to
// hook the check on other conditions. It should be used IN said permission
// hooks.
//
// If ch is non-nil, this is a channel op flag check and the user's membership
// entry on the channel will be checked. If no such entry exists, the check
// automatically fails. If ch is nil, this is a server op flag check, and the user's own metadata will be checked.
func HasOpFlag(u *core.User, ch *core.Channel, flag string) bool {
	var e core.Extensible
	var defwords []string
	if ch == nil {
		e = u
		defwords = defServerOp
	} else {
		if m := ch.GetMember(u); m != nil {
			e = m
		} else {
			return false
		}
		defwords = defChanOp
	}

	words := strings.Fields(e.Data("op"))
	for _, word := range words {
		if word == flag {
			return true
		}
		if word == "on" {
			for _, defword := range defwords {
				if defword == flag {
					return true
				}
			}
		}
	}

	return false
}
