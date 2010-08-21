/*
	
*/
package perm

import "strings"

import "oddircd/src/core"


var banTypes map[string]func(u *core.User, mask string)bool

var defaultBan map[string]string

func init() {
	banTypes = make(map[string]func(u *core.User, mask string)bool)
	defaultBan = make(map[string]string)

	defaultBan["ban"] = "join mute nick"
}


// AddBanType adds the given ban type handler to the ban type map.
// The handler will receive the user and mask, and must return a bool
// indicating whether the user matched the ban or not.
// This may only be used during init.
func AddBanType(t string, h func(u *core.User, mask string) bool) {
	banTypes[t] = h
}

// DefaultBan returns the default list of restrictions for the given ban list.
// They are space-separated.
func DefaultBan(list string) string {
	if def := defaultBan[list]; def != "" {
		return def
	}
	
	return defaultBan["ban"]
}

// AddDefaultBan adds the given restriction to the default ban restrictions for
// the given ban list.
// It may be used only during init.
func AddDefaultBan(list, name string) {
	defaultBan[list] = defaultBan[list] + " " + name
}

// BanMatches returns whether the user has a ban on the given Extensible,
// matching the given restriction, under the the given prefix for ban types.
// It handles checking for default bans, if this ban type is part of the
// default.
func BanMatch(u *core.User, e core.Extensible, prefix, restrict string) (match bool) {
	def := defaultBan[prefix]
	if def == "" {
		def = defaultBan["ban"]
	}
	defwords := strings.Fields(def)

	for t, h := range banTypes {
		prefixlen := len(prefix) + len(t) + 2
		e.DataRange(prefix + " " + t + " ",
		func(name, value string) {
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

			if h(u, name[prefixlen:]) {
				match = true
			}
		})

		if match == true {
			break
		}
	}

	return
}
