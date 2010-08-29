package irc

import "fmt"
import "os"
import "strings"
import "time"

import "oddcomm/src/core"


// ParseModeLine parses a line of mode changes into core.DataChange structs.
// Redundant changes are compressed down into one.
// An error is returned if unknown modes are encountered, or modes are dropped
// due to missing parameters. The remainder of the modes are still parsed.
// e is the user or channel being changed.
func (p *ModeParser) ParseModeLine(source *core.User, e core.Extensible, modeline []byte, params [][]byte) (*core.DataChange, os.Error) {
	var adding bool = true
	var unknown string
	var missing string
	var param int
	modes := string(modeline)
	changes := make (map[string]*core.DataChange)

	for _, char := range modes {
		if char == '+' {
			adding = true
			continue
		} else if char == '-' {
			adding = false
			continue
		}
	
		if v, ok := p.simple[char]; ok {
			if v, ok := p.extended[char]; ok {
				newchanges := v(adding, e, "")
				for it := newchanges; it != nil; it = it.Next {
					changes[it.Name] = it
				}
				continue
			}
	
			change := new(core.DataChange)
			change.Name = v
			if adding {
				change.Data = "on"
			}

			changes[change.Name] = change	
			continue
		}

		if v, ok := p.parametered[char]; ok {
			var change *core.DataChange
			if adding {
				if param >= len(params) {
					missing += string(char)
					continue
				}

				if v, ok := p.extended[char]; ok {
					newchanges := v(adding, e, string(params[param]))
					param++
					for it := newchanges; it != nil
							it = it.Next {
						changes[it.Name] = it
					}
					continue
				}

				change = new(core.DataChange)
				change.Name = v
				change.Data = string(params[param])
				param++
			} else {
				if v, ok := p.extended[char]; ok {
					newchanges := v(adding, e, "")
					for it := newchanges; it != nil
							it = it.Next {
						changes[it.Name] = it
					}
					continue
				}

				change = new(core.DataChange)
				change.Name = v
			}

			changes[change.Name] = change
			continue
		}
		
		if v, ok := p.list[char]; ok {
			if param >= len(params) {
				missing += string(char)
				continue
			}
			change := new(core.DataChange)
			cparam := string(params[param])
			param++

			if v, ok := p.extended[char]; ok {
				newchanges := v(adding, e, cparam)
				for it := newchanges; it != nil; it = it.Next {
					if it.Data != "" {
						it.Data += fmt.Sprintf(" setby-%s setat-%d", source.GetSetBy(), time.Seconds())
					}
					changes[it.Name] = it
				}
				continue
			}

			if adding {
				change.Name = v + " " + cparam
				change.Data = fmt.Sprintf("on setby-%s setat-%d", source.GetSetBy(), time.Seconds())
			} else {
				change.Name = v + " " + cparam
			}

			changes[change.Name] = change
			continue
		}

		if v, ok := p.membership[char]; ok {
			var ch *core.Channel
			var ok bool
			if ch, ok = e.(*core.Channel); !ok {
				continue
			}
			if param >= len(params) {
				missing += string(char)
				continue
			}
			par := string(params[param])
			param++

			if v, ok := p.extended[char]; ok {
				newchanges := v(adding, e, par)
				for it := newchanges; it != nil; it = it.Next {
					if it.Member != nil {
						changes["m" + it.Member.User().ID() + " " + it.Name] = it
					} else {
						changes[it.Name] = it
					}
				}
				continue
			}

			var u *core.User
			var m *core.Membership
			if u = core.GetUser(par); u == nil {
				if u = core.GetUserByNick(par); u == nil {
					continue
				}
			}
			if m = ch.GetMember(u); m == nil {
				continue
			}
			
			change := new(core.DataChange)
			change.Name = v
			change.Member = m

			if adding {
				change.Data = "on"
			}

			changes["m" + change.Member.User().ID() + " " + change.Name] = change
			continue
		}

		unknown += string(char)
	}

	// Turn the modes into a list.
	var c *core.DataChange
	for change := range changes {
		changes[change].Next = c
		c = changes[change]
	}

	// Get the error, if we had one.
	var errstring string
	if unknown != "" {
		errstring += "Unknown mode letters: " + unknown
		if missing != "" {
			errstring += "  "
		}
	}
	if missing != "" {
		errstring += "Missing parameters for: " + missing
	}
	var err os.Error
	if errstring != "" {
		err = os.NewError(errstring)
	}

	return c, err
}

// ParseChanges parses a list of mode changes into a line of mode changes and
// parameters. Changes which do not correspond to a mode are dropped.
// e is the user or channel being changed.
func (p *ModeParser) ParseChanges(e core.Extensible, c *core.DataChange,
                                  old *core.OldData) (modeline string) {
	var addmodes string
	var remmodes string
	var addparams string
	var remparams string

	for it, o := c, old; it != nil && o != nil; it, o = it.Next, o.Next {
		if v, ok := p.nameToExt[it.Name]; ok && v != nil {
			add, addpar, rem, rempar := v(e, it.Name, o.Data, it.Data)
			addmodes += string(add)
			remmodes += string(rem)
			for _, par := range addpar {
				addparams += " " + par
			}
			for _, par := range rempar {
				remparams += " " + par
			}
			continue
		}
	
		if v, ok := p.nameToSimple[it.Name]; ok {
			if it.Data != "" {
				addmodes += string(v)
			} else {
				remmodes += string(v)
			}
			continue
		}

		if v, ok := p.nameToParametered[it.Name]; ok {
			if it.Data != "" {
				addmodes += string(v)
				addparams += " " + it.Data
			} else {
				remmodes += string(v)
			}
			continue
		}

		if v, ok := p.nameToMembership[it.Name]; ok {
			if it.Member == nil {
				// Wha
				continue
			}

			var param string
			if p.uids {
				param = it.Member.User().ID()
			} else {
				param = it.Member.User().Nick()
			}

			if it.Data != "" {
				addmodes += string(v)
				addparams += " " + param
			} else {
				remmodes += string(v)
				remparams += " " + param
			}
			continue
		}

		var subentry = strings.LastIndex(it.Name, " ") + 1
		for subentry > 2 {
			if v, ok := p.nameToExt[it.Name[0:subentry-1]]; ok &&
					v != nil {
				add, addpar, rem, rempar := v(e, it.Name, o.Data, it.Data)
				addmodes += string(add)
				remmodes += string(rem)
				for _, par := range addpar {
					addparams += " " + par
				}
				for _, par := range rempar {
					remparams += " " + par
				}
				break
			}

			if v, ok := p.nameToList[it.Name[0:subentry-1]]; ok {
				if it.Data != "" {
					addmodes += string(v)
					addparams += " " + it.Name[subentry:]
				} else {
					remmodes += string(v)
					remparams += " " + it.Name[subentry:]
				}
				break
			}
			subentry = strings.LastIndex(it.Name[subentry:], " ") +
			           1
		}
	}

	var result string
	if addmodes != "" {
		result += "+" + addmodes
	}
	if remmodes != "" {
		result += "-" + remmodes
	}
	return result + addparams + remparams
}


// GetModes gets the modeline associated with a user or channel.
// It caches its result via metadata on the user, using the mode parser's
// mode ID to disambiguate it from other mode parsers.
func (p* ModeParser) GetModes(e core.Extensible) string {
	var modes string
	var params string

	for name, char := range p.nameToSimple {
		data := e.Data(name)

		if v, ok := p.getExt[char]; ok {
			data = v(e)
		}

		if data != "" {
			modes += string(char)
		}
	}

	for name, char := range p.nameToParametered {
		data := e.Data(name)

		if v, ok := p.getExt[char]; ok {
			data = v(e)
		}

		if data != "" {
			modes += string(char)
			params += " " + data
		}
	}

	return modes + params
}


// GetModeList calls the given function for every entry in the list mode,
// with the parameter of this mode list entry, and its metadata value. If
// there are none, it will not be called.
func (p* ModeParser) ListMode(e core.Extensible, char int, f func(param, value string)) bool {
	prefix := p.list[char]
	if prefix != "" {
		e.DataRange(prefix + " ", func(name, value string) {
			if v, ok := p.nameToExt[prefix]; ok {
				_, addpar, _, _ := v(e, name, "", value)
				for _, par := range addpar {
					f(par, value)
				}
			} else {
				f(name[len(prefix)+1:], value)
			}
		})
		return true
	}
	return false
}
