package irc

import "fmt"
import "os"
import "strings"
import "time"

import "oddircd/src/core"

var currentID uint64

// Stores a mapping of codepoint character modes to metadata strings they
// correspond to.
type ModeParser struct {
	id uint64
	uids bool
	simple map[int]string
	parametered map[int]string
	extended map[int]func(bool, core.Extensible, string)(*core.DataChange)
	list map[int]string
	membership map[int]string
	nameToSimple map[string]int
	nameToParametered map[string]int
	nameToList map[string]int
	nameToMembership map[string]int
	nameToExt map[string]func(core.Extensible, string, string, string)([]int, []string, []int, []string)
	getExt map[int]func(core.Extensible)string
	prefixes *prefix
}

// NewModeParser returns a new mode parser, ready to add modes to.
// uids indicates whether this mode parser should output UIDs or nicks
// for membership changes, and is only relevant if this mode parser is going
// to be used for channel modes.
// This method cannot be called more than 2^64-1 times or it runs out of IDs.
// And then it's game over, man, game over.
func NewModeParser(uids bool) (p *ModeParser) {
	p = new(ModeParser)
	p.uids = uids

	// Set our ID.
	p.id = currentID; currentID++

	// Initialise ALL our maps.
	p.simple = make(map[int]string)
	p.parametered = make(map[int]string)
	p.list = make(map[int]string)
	p.membership = make(map[int]string)
	p.extended = make(map[int]func(bool, core.Extensible, string)(*core.DataChange))
	p.nameToSimple = make(map[string]int)
	p.nameToParametered = make(map[string]int)
	p.nameToList = make(map[string]int)
	p.nameToMembership = make(map[string]int)
	p.nameToExt = make(map[string]func(core.Extensible, string, string, string)([]int, []string, []int, []string))
	p.getExt = make(map[int]func(core.Extensible)string)

	// Add the base prefixes.
	p.AddPrefix('@', "op", 100000)
	p.AddPrefix('+', "voiced", 100)

	return
}

// AddSimple adds a simple mode. It cannot be called concurrently with itself,
// or any lookups on the parser.
// If a mode character or metadata name is added twice, the mode associated
// with the previous value is deleted.
func (p *ModeParser) AddSimple(mode int, metadata string) {
	if v := p.simple[mode]; v != "" {
		p.nameToSimple[v] = 0, false
	}
	if v := p.nameToSimple[metadata]; v != 0 {
		p.simple[v] = "", false
	}
	p.simple[mode] = metadata
	p.nameToSimple[metadata] = mode
}

// AddParametered adds a parametered mode. It cannot be called concurrently
// with itself, or any lookups on the parser.
// If a mode character or metadata name is added twice, the mode associated
// with the previous value is deleted.
func (p *ModeParser) AddParametered(mode int, metadata string) {
	if v := p.parametered[mode]; v != "" {
		p.nameToParametered[v] = 0, false
	}
	if v := p.nameToParametered[metadata]; v != 0 {
		p.parametered[v] = "", false
	}
	p.parametered[mode] = metadata
	p.nameToParametered[metadata] = mode
}

// AddList adds a list mode. It cannot be called concurrently with itself, or
// any lookups on the parser.
// If a mode character or metadata name is added twice, the mode associated
// with the previous value is deleted.
func (p *ModeParser) AddList(mode int, metadata string) {
	if v := p.list[mode]; v != "" {
		p.nameToList[v] = 0, false
	}
	if v := p.nameToList[metadata]; v != 0 {
		p.list[v] = "", false
	}
	p.list[mode] = metadata
	p.nameToList[metadata] = mode
}

// AddMembership adds a membership mode. It cannot be called concurrently with
// itself, or any lookups on the parser.
// If a mode character or metadata name is added twice, the mode associated
// with the previous value is deleted.
// Membership modes are ignored if this is not a channel mode parser.
func (p *ModeParser) AddMembership(mode int, metadata string) {
	if v := p.membership[mode]; v != "" {
		p.nameToMembership[v] = 0, false
	}
	if v := p.nameToMembership[metadata]; v != 0 {
		p.membership[v] = "", false
	}
	p.membership[mode] = metadata
	p.nameToMembership[metadata] = mode
}

// AddExtMode extends an already added mode, attaching hooks to mapping the
// mode to metadata, its metadata to modes, and generating a list of set modes
// on a user. It permits modes which require additional logic specific to that
// mode.
// The name will be treated as a prefix; both it directly, and all subentries,
// will be fed to the nameToMode function when parsing changes.
//
// The mode must already have been added to one of the other types; this will
// determine whether its hooks receive a parameter on set (parametered),
// set and unset (list), or never (simple). It also determines whether the
// getSet hook will be called (simple or parametered only), and whether it may
// return a parameter to be listed.

// It is legal to have the metadata name be nil, in which case the mode will
// not map to any specific name, and nameToMode will never be called and may
// be nil. A placeholder metadata name must still be given while adding it as
// one of the normal types.
//
// The modeToName function should expect a boolean indicating whether the mode
// is being added or removed, and the user/channel being changed followed by
// the mode's parameter. It must return a list of DataChange objects to apply,
// which may be empty.
//
// The nameToMode function shuld expect the user/channel being changed
// followed by the metadata item's full name, its previous value, and its
// current value. It should return a slice of added mode characters and
// another of removed mode characters, and a slice of added mode parameters
// and another of removed mode parameters. Any of these may be empty, and the
// added or removed mode characters may include modes other than the mode
// which corresponds to this name if and only if the previous value is not "".
// This function will also be called to convert existing metadata into list
// mode entries, with a previous value of "".
//
// The getSet function should expect a user or channel, and return a value for
// the mode, which may be "" to indicate it is unset, or any other value to
// indicate it is set, and for parametered modes, with that as the parameter.
//
// This method cannot be called concurrently with itself, or any lookups on
// the parser.
func (p *ModeParser) AddExtMode(mode int, name string, modeToName func(bool, core.Extensible, string) (*core.DataChange), nameToMode func(core.Extensible, string, string, string) ([]int, []string, []int, []string), getSet func(core.Extensible) string) {
	p.extended[mode] = modeToName
	p.getExt[mode] = getSet
	if name != "" {
		p.nameToExt[name] = nameToMode
	}
}


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

		var subentry = strings.IndexRune(it.Name, ' ') + 1
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
			subentry = strings.IndexRune(it.Name[subentry:], ' ') +
			           subentry + 1
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
