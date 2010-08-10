package irc

import "os"
import "strings"

import "oddircd/src/core"


// Stores a mapping of codepoint character modes to metadata strings they
// correspond to.
type ModeParser struct {
	simple map[int]string
	parametered map[int]string
	extended map[int]func(bool, core.Extensible, string)(*core.DataChange)
	list map[int]string
	nameToSimple map[string]int
	nameToParametered map[string]int
	nameToList map[string]int
	nameToExt map[string]func(core.Extensible, string, string)([]int, []string, []int, []string)
}

// NewModeParser returns a new mode parser, ready to add modes to.
// channel sets whether this is a channel mode parser, or user mode parser.
// This determines which modes to fake the existence of for compatibility.
func NewModeParser() (p *ModeParser) {
	p = new(ModeParser)
	p.simple = make(map[int]string)
	p.parametered = make(map[int]string)
	p.list = make(map[int]string)
	p.extended = make(map[int]func(bool, core.Extensible, string)(*core.DataChange))
	p.nameToSimple = make(map[string]int)
	p.nameToParametered = make(map[string]int)
	p.nameToList = make(map[string]int)
	p.nameToExt = make(map[string]func(core.Extensible, string, string)([]int, []string, []int, []string))
	return
}

// AddSimple adds a simple mode. It cannot be called concurrently with itself,
// or any lookups on the parser.
func (p *ModeParser) AddSimple(mode int, metadata string) {
	p.simple[mode] = metadata
	p.nameToSimple[metadata] = mode
}

// AddParametered adds a parametered mode. It cannot be called concurrently
// with itself, or any lookups on the parser.
func (p *ModeParser) AddParametered(mode int, metadata string) {
	p.parametered[mode] = metadata
	p.nameToParametered[metadata] = mode
}

// AddList adds a list mode. It cannot be called concurrently with itself, or
// any lookups on the parser.
func (p *ModeParser) AddList(mode int, metadata string) {
	p.list[mode] = metadata
	p.nameToList[metadata] = mode
}

// AddExtMode adds a mode/metadata name combination which requires additional
// logic to map a mode change to metadata, and visa versa.
// The name will be treated as a prefix; both it directly, and all subentries,
// will be fed to the nameToMode function when parsing changes.
//
// The mode must also be added to one of the other types; this will determine
// whether it receives a parameter on set (parametered), set and unset (list),
// or never (simple). It is legal to have either the metadata name or the mode
// be nil to extend the mode only one way, or create a mode that does not
// correspond to any single piece of metadata, but a placeholder entry is
// still needed in the appropriate other type. It is in this case permitted
// to provide a nil function for the unused mapping.
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
// and another of removed mode parameters. Any of these may be empty.
//
// This method cannot be called concurrently with itself, or any lookups on
// the parser.
func (p *ModeParser) AddExtMode(mode int, name string, modeToName func(bool, core.Extensible, string) (*core.DataChange), nameToMode func(core.Extensible, string, string) ([]int, []string, []int, []string)) {
	if mode != 0 {
		p.extended[mode] = modeToName
	}
	if name != "" {
		p.nameToExt[name] = nameToMode
	}
}


// ParseModeLine parses a line of mode changes into core.DataChange structs.
// Redundant changes are compressed down into one.
// An error is returned if unknown modes are encountered, or modes are dropped
// due to missing parameters. The remainder of the modes are still parsed.
// e is the user or channel being changed.
func (p *ModeParser) ParseModeLine(e core.Extensible, modeline []byte, params [][]byte) (*core.DataChange, os.Error) {
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

			changes[v] = change	
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

			changes[v] = change
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
					changes[it.Name] = it
				}
				continue
			}

			if adding {
				change.Name = v + " " + cparam
				change.Data = "on"
			} else {
				change.Name = v + " " + cparam
			}

			changes[v] = change
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

	for it, o := c, old; it != nil && o != nil; it, o = c.Next, o.Next {
		if v, ok := p.nameToExt[it.Name]; ok {
			add, addpar, rem, rempar := v(e, o.Data, it.Data)
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

		var subentry = strings.IndexRune(it.Name, ' ') + 1
		if subentry != 0 {
			if v, ok := p.nameToList[it.Name[0:subentry-1]]; ok {
				if it.Data != "" {
					addmodes += string(v)
					addparams += " " + it.Data
				} else {
					remmodes += string(v)
					remparams += " " + it.Data
				}
				continue
			}
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
