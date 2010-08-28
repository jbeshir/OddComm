/*
	Implements extended ban types, exporting methods for other modules
	to add their own.

	Includes extended bans for the default restrictions and ban types
	supported by the core.
*/
package extbans

import "os"
import "strings"
import "utf8"

import "oddcomm/src/client"
import "oddcomm/src/core"
import "oddcomm/lib/irc"
import "oddcomm/lib/perm"

var MODULENAME string = "modules/client/extbans"


// ExtBanRestrict is the mapper for extban characters to ban restrictions.
var ExtBanRestrict *irc.CMapper

// ExtBanType is the map of extban characters to ban type.
var ExtBanType *irc.CMapper


func init() {
	ExtBanRestrict = irc.NewCMapper()
	ExtBanType = irc.NewCMapper()

	// Add the built-in ban types.
	ExtBanType.Add('H', "host")
	ExtBanType.Add('A', "account")

	// Add the built-in ban restrictions.
	ExtBanRestrict.Add('j', "join")
	ExtBanRestrict.Add('m', "mute")
	ExtBanRestrict.Add('n', "nick")

	// Add ban, ban exception, and unrestrict/invex modes on channels.
	// This immediately overrides the core channel modes.
	client.ChanModes.AddList('b', "ban")
	client.ChanModes.AddList('e', "banexception")
	client.ChanModes.AddList('I', "unrestrict")
	
	// Extend ban mode.
	client.ChanModes.ExtendModeToData('b', func(adding bool, e core.Extensible, param string) *core.DataChange {
		return processBan("ban", perm.DefaultBan(), adding, e, param)
	})
	client.ChanModes.ExtendDataToMode("ban", func (e core.Extensible, name, oldvalue, newvalue string) ([]int, []string, []int, []string) {
		return makeBan('b', perm.DefaultBan(), e, name, oldvalue,
		               newvalue)
	})

	// Extend ban exception mode.
	client.ChanModes.ExtendModeToData('e', func(adding bool, e core.Extensible, param string) *core.DataChange {
		return processBan("banexception", perm.DefaultBan(), adding, e,
		                  param)
	})
	client.ChanModes.ExtendDataToMode("banexception", func (e core.Extensible, name, oldvalue, newvalue string) ([]int, []string, []int, []string) {
		return makeBan('e', perm.DefaultBan(), e, name, oldvalue,
		               newvalue)
	})

	// Extend unrestrict (invex) mode.
	client.ChanModes.ExtendModeToData('I', func(adding bool, e core.Extensible, param string) *core.DataChange {
		return processBan("unrestrict", perm.DefaultBan(), adding, e,
		                  param)
	})
	client.ChanModes.ExtendDataToMode("unrestrict", func (e core.Extensible, name, oldvalue, newvalue string) ([]int, []string, []int, []string) {
		return makeBan('I', perm.DefaultBan(), e, name, oldvalue,
		               newvalue)
	})
	
	// Block colons from use in nicks and idents.
	perm.HookCheckNick(func(_ *core.User, nick string) (int, os.Error) {
		if strings.IndexRune(nick, ':') != -1 {
			return -1e9, os.NewError("Nick contains colon.")
		}
		return 0, nil
	})
	perm.HookCheckUserData("ident", func(_, _ *core.User, _, ident string) (int, os.Error) {
		if strings.IndexRune(ident, ':') != -1 {
			return -1e9, os.NewError("Ident contains colon.")
		}
		return 0, nil
	})
}

// Function handling processing of ban syntax into metadata.
// This is called to handle bans, ban exceptions, and unrestrictions, and
// returns the data change object, sans the prefix for the metadata name.
func processBan(prefix, def string, adding bool, e core.Extensible, param string) *core.DataChange {
	var change core.DataChange
	change.Data = def
	t := "host"
	defaultType := true

	// Find colons, indicating an extban.
	var first, second, mask int
	first = strings.IndexRune(param, ':')
	if first > -1 && len(param) > first+1 {
		mask = first + 1
		second = first + 1 + strings.IndexRune(param[first+1:], ':')
		if second == first {
			second = -1
		}
		if second != -1 && len(param) > second + 1 {
			mask = second + 1
		}
	}

	// If both colons exist, treat everything before the first as
	// restrictions, and the first character in the second as the ban
	// type.
	if first > -1 && second > -1 {
		change.Data = ""
		for _, char := range param[0:first] {
			if v := ExtBanRestrict.Str(char); v != "" {
				if change.Data != "" {
					change.Data += " "
				}
				change.Data += v
			}
		}

		var reverse bool
		char, _ := utf8.DecodeRuneInString(param[first+1:])
		if char == '~' {
			reverse = true
			char, _ = utf8.DecodeRuneInString(param[first+2:])
		}
		if v := ExtBanType.Str(char); v != "" {
			t = v
			if reverse {
				t = "~" + t
			}
			defaultType = false
		}
	} else if first > -1 {
		// If we just got one colon, then if there is only a single
		// character before it (possibly with a ~), and it is a valid
		// ban type, interpret it as a type.
		var exttype string
		var reverse bool
		char, pos := utf8.DecodeRuneInString(param)
		if char == '~' {
			reverse = true
			char, pos = utf8.DecodeRuneInString(param[1:])
			pos += 1
		}
		if param[pos] == ':' {
			if v := ExtBanType.Str(char); v != "" {
				exttype = v
				if reverse {
					exttype = "~" + exttype
				}
				t = exttype
				defaultType = false
			}
		}

		// Otherwise, interpret things before as restrictions.
		if exttype == "" {
			change.Data = ""
			for _, char := range param[0:first] {
				if v := ExtBanRestrict.Str(char); v != "" {

					// Omit duplicates.
					words := strings.Fields(change.Data)
					var alreadyThere bool
					for _, word := range words {
						if v == word {
							alreadyThere = true
							break
						}
					}
					if alreadyThere {
						continue
					}

					if change.Data != "" {
						change.Data += " "
					}
					change.Data += v
				}
			}
		}
	}

	// Set the name, now we have the type and mask separated.
	change.Name = prefix + " " + t + " " + param[mask:]

	// If we've on the host type, automatically expand the host if they
	// provided a partial one.
	if t == "host" {
		if strings.IndexRune(param[mask:], '!') == -1 {

			// Default bans: If no explicit type was given, and
			// just a valid user nickname is given, provide a good
			// default ban for the user.
			if defaultType {
				u := core.GetUserByNick(param[mask:])
				if u != nil {
					change.Name = prefix + " " +
					              u.GetDecentBan()
				} else {
					change.Name += "!*@*"
				}
			} else {
				change.Name += "!*@*"
			}
		} else if strings.IndexRune(param[mask:], '@') == -1 {
			change.Name += "@*"
		}
	}

	// Get the existing restrictions, expanding default restrictions.
	existingData := e.Data(change.Name)
	words := strings.Fields(existingData)
	var existing string
	for _, word := range words {
		if len(word) > 5 && word[0:5] == "setby" {
			continue
		}
		if len(word) > 5 && word[0:5] == "setat" {
			continue
		}

		if word == "on" {
			word = def
		}
		
		if existing != "" {
			existing += " "
		}
		existing += word
	}
		
	if adding {
		// If we're adding the ban, add the new restrictions to the
		// previous restrictions.
		if existing != "" {
			words := strings.Fields(existing)
			remwords := strings.Fields(change.Data)
			for _, restrict := range words {
				var found bool
				for _, w := range remwords {
					if w == restrict {
						found = true
						break
					}
				}
				if !found {
					if change.Data != "" {
						change.Data += " "
					}
					change.Data += restrict
				}
			}
		}
	} else {
		// If we're removing the ban, remove the removed restrictions.
		// Leave restrictions not removed alone.
		// This is "fun". This is also O(n^2).
		var left string
		if existing != "" {
			words := strings.Fields(existing)
			remwords := strings.Fields(change.Data)
			for _, restrict := range words {
				var found bool
				for _, w := range remwords {
					if w == restrict {
						found = true
						break
					}
				}
				if !found && ExtBanRestrict.Char(restrict) != 0 {
					if left != "" {
						left += " "
					}
					left += restrict
				}
			}
		}
		change.Data = left
	}

	// Test for whether we have a default restriction.
	var outsideDefault bool
	var missingDefault bool
	defwords := strings.Fields(def)
	words = strings.Fields(change.Data)
	for _, word := range words {
		var found bool
		for _, defword := range defwords {
			if defword == word {
				found = true
				break
			}
		}
		if !found {
			outsideDefault = true
			break
		}
	}
	if !outsideDefault {
		for _, defword := range defwords {
			var found bool
			for _, word := range words {
				if defword == word {
					found = true
					break
				}
			}
			if !found {
				missingDefault = true
				break
			}
		}
	}

	// If the restriction is the same as the default restriction, quietly
	// change it to "on".
	if !outsideDefault && !missingDefault {
		change.Data = "on"
	}

	return &change
}

// Function handling processing of ban metadata changes into ban syntax.
// This is called to handle bans, ban exceptions, and unrestrictions, and
// returns values as suitable for passing as the result of nameToMode.
func makeBan(char int, def string, e core.Extensible, name, oldvalue, newvalue string) (add []int, addparam []string, rem []int, remparam []string) {

	// Get the new ban parameter, and the old ban parameter.
	var newparam string
	if newvalue != "" {
		newparam = makeBanParam(def, name, newvalue)
	}
	var oldparam string
	if oldvalue != "" {
		oldparam = makeBanParam(def, name, oldvalue)
	}

	// If the ban parameters match, return; no user-visible change.
	if newparam == oldparam {
		return
	}

	// If we have a new ban, set it.
	if newparam != "" {
		add = []int{char}
		addparam = new([1]string)
		addparam[0] = newparam
	}

	// If we had an old ban, unset it.
	if oldparam != "" {
		rem = []int{char}
		remparam = new([1]string)
		remparam[0] = oldparam
	}

	return
}

// Makes a ban parameter from the ban name, value, and default restrictions.
func makeBanParam(def, name, value string) (param string) {
	// If this ban's name is too short, ignore it.
	namewords := strings.Fields(name)
	if len(namewords) < 3 {
		return
	}

	// Look up the ban's type character, returning if it has none.
	// If "host", don't have a type string; it's a default ban.
	t := namewords[1]
	var tstring string
	if t != "host" {
		if t[0] == '~' {
			t = t[1:]
			tstring = "~"
		}
		if v := ExtBanType.Char(t); v != 0 {
			tstring += string(v) + ":"
		} else {
			return
		}
	}

	// Look up the ban's restrict characters. Ones we don't recognise are
	// ignored. Remember whether we found any which were outside the
	// default set.
	restrictions := ""
	valwords := strings.Fields(value)
	defwords := strings.Fields(def)
	var outsideDefault bool
	for _, word := range valwords {
		if word == "on" {
			for _, word := range defwords {
				if v := ExtBanRestrict.Char(word); v != 0 {
					restrictions += string(v)
				}
			}
			continue
		}
		if v := ExtBanRestrict.Char(word); v != 0 {
			outsideDefault = true
			restrictions += string(v)
		}
	}

	// If there are no known restriction characters, return; nothing this
	// ban does is visible.
	if restrictions == "" {
		return
	}
	
	// If we have the default restrictions, clear the restrictions string.
	// Otherwise, append a colon to it, it's done.
	if !outsideDefault {
		restrictions = ""
	} else {
		restrictions += ":"
	}

	// Put the pieces of the ban string together.
	param = restrictions + tstring + namewords[2]
	return
}
