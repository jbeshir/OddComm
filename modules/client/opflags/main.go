/*
	Implements an extended syntax for granting op flags when setting a
	user +o, in the form "+o <flags>:<nick>"
*/
package opflags

import "os"
import "strings"

import "oddircd/src/client"
import "oddircd/src/core"
import "oddircd/src/irc"
import "oddircd/src/perm"

var MODULENAME string = "modules/client/opflags"


// Flags is the mapper of op flag characters to op flags.
var Flags *irc.CMapper


func init() {
	Flags = irc.NewCMapper()

	// Add opflag characters for all the default core opflags.
	Flags.Add('b', "ban")
	Flags.Add('c', "banview")
	Flags.Add('i', "invite")
	Flags.Add('m', "moderate")
	Flags.Add('p', "msg")
	Flags.Add('t', "topic")
	Flags.Add('o', "(de)op")
	Flags.Add('v', "(de)mark")

	// Extend op mode.
	client.ChanModes.AddExtMode('o', "op",
	                     func(adding bool, e core.Extensible,
	                          param string) *core.DataChange {
		ch, ok := e.(*core.Channel)
		if !ok {
			return nil
		}
		return processOp(adding, ch, param)
	}, nil , nil)

	// Block colons from use in nicks and idents.
	NoColons := func(u *core.User, nick string) (int, os.Error) {
		if strings.IndexRune(nick, ':') != -1 {
			return -1e9, os.NewError("Parameter contains colon.")
		}
		return 0, nil
	}
	perm.HookValidateNick(NoColons)
	perm.HookValidateIdent(NoColons)
}

// Function handling processing of extended op syntax into metadata.
// It returns the data change object.
func processOp(adding bool, ch *core.Channel, param string) *core.DataChange {
	var change core.DataChange
	change.Name = "op"
	change.Data = perm.ChanDefaultOp()

	// Find a colon, indicating extended op syntax.
	var colon, mask int
	colon = strings.IndexRune(param, ':')
	if colon > -1 && len(param) > colon+1 {
		mask = colon + 1
	} else {
		colon = -1
	}

	// Set the member this opping refers to.
	if target := core.GetUserByNick(param[mask:]); target != nil {
		if m := ch.GetMember(target); m != nil {
			change.Member = m
		} else {
			return nil
		}
	} else {
		return nil
	}

	// If a colon exists, treat everything before it as opflags.
	if colon > -1 {
		change.Data = ""
		for _, char := range param[0:colon] {
			if v := Flags.Str(char); v != "" {
				if change.Data != "" {
					change.Data += " "
				}
				change.Data += v
			}
		}
	}

	// Get the existing op flags, expanding default ops.
	existingData := change.Member.Data("op")
	words := strings.Fields(existingData)
	var existing string
	for _, word := range words {
		if word == "on" {
			word = perm.ChanDefaultOp()
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
			for _, flag := range words {
				var found bool
				for _, w := range remwords {
					if w == flag {
						found = true
						break
					}
				}
				if !found {
					if change.Data != "" {
						change.Data += " "
					}
					change.Data += flag
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
			for _, flag := range words {
				var found bool
				for _, w := range remwords {
					if w == flag {
						found = true
						break
					}
				}
				if !found && Flags.Char(flag) != 0 {
					if left != "" {
						left += " "
					}
					left += flag
				}
			}
		}
		change.Data = left
	}

	// Test for whether we have a default restriction.
	var outsideDefault bool
	var missingDefault bool
	defwords := strings.Fields(perm.ChanDefaultOp())
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
