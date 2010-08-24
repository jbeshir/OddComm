/*
	
*/
package perm

import "strings"

import "oddircd/src/core"


var banTypes map[string]func(u *core.User, mask string)bool

var defaultBan string = "join mute nick"
var defaultUnrestrict string = "join"

func init() {
	banTypes = make(map[string]func(u *core.User, mask string)bool)
}


// AddBanType adds the given ban type handler to the ban type map.
// The handler will receive the user and mask, and must return a bool
// indicating whether the user matched the ban or not.
// This may only be used during init.
func AddBanType(t string, h func(u *core.User, mask string) bool) {
	banTypes[t] = h
}

// DefaultBan returns the default list of restrictions for a default ban.
func DefaultBan() string {
	return defaultBan
}

// DefaultUnrestrict returns the default list of restrictions for a default
// unrestriction.
func DefaultUnrestrict() string {
	return defaultUnrestrict
}

// AddDefaultBan adds the given restriction to the default ban restrictions.
// It may be used only during init.
func AddDefaultBan(name string) {
	defaultBan += " " + name
}

// AddDefaultUnrestrict adds the given restriction to the default
// unrestrictions.
// It may be used only during init.
func AddDefaultUnrestrict(name string) {
	defaultUnrestrict += " " + name
}

// Banned returns whether the user has a ban on the given Extensible,
// matching the given restriction. It handles checking for default bans, if
// this ban type is part of the default, and ban exceptions.
func Banned(u *core.User, e core.Extensible, restrict string) bool {
	if banMatch(u, e, "ban", restrict) {
		if !banMatch(u, e, "banexception", restrict) {
			return true
		}
	}
	return false
}

// Restricted returns whether the user is restricted by a restriction of the
// given type on the given extensible. It handles checking for unrestriction
// modes, as well as the given restriction being set.
func Restricted(u *core.User, e core.Extensible, restrict string) bool {
	if e.Data("restrict " + restrict) != "" {
		if !banMatch(u, e, "unrestrict", restrict) {
			return true
		}
	}
	return false
}

// banMatch returns whether the user has a ban on the given Extensible,
// matching the given restriction, under the given prefix for ban types.
// It handles checking for default bans, if this ban type is part of the
// default.
func banMatch(u *core.User, e core.Extensible, prefix, restrict string) (match bool) {
	def := defaultBan
	if prefix == "unrestrict" {
		def = defaultUnrestrict
	}
	defwords := strings.Fields(def)

	// Create matcher closure, to run for each ban.
	var reverse bool
	var prefixlen int
	var t string
	var hook func(u *core.User, mask string)bool
	matcher := func(name, value string) {
		words := strings.Fields(value)
		var found bool
		for _, word := range words {
			if word == restrict {
				found = true
				break
			}
			if word == "on" {
				var deffound bool
				for _, defword := range defwords {
					if defword == restrict {
						deffound = true
						break
					}
				}
				if deffound {
					found = true
					break
				}
			}
		}

		if !found {
			return
		}

		// If the hook returns true and we're NOT reversed, match.
		// If the hook returns false and we are reversed, match.
		if hook(u, name[prefixlen:]) == !reverse {
			match = true
		}
	}

	for t, hook = range banTypes {
		reverse = false
		prefixlen = len(prefix) + len(t) + 2
		e.DataRange(prefix + " " + t + " ", matcher)
		if match == true {
			break
		}

		reverse = true
		prefixlen = len(prefix) + len(t) + 3
		e.DataRange(prefix + " ~" + t + " ", matcher)
		if match == true {
			break
		}
	}

	return
}
