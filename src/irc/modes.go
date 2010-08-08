package irc

import "os"
import "strings"

import "oddircd/src/core"


// Stores a mapping of codepoint character modes to metadata strings they
// correspond to.
type ModeParser struct {
	simple map[int]string
	parametered map[int]string
	list map[int]string
	nameToSimple map[string]int
	nameToParametered map[string]int
	nameToList map[string]int
}

// NewModeParser returns a new mode parser, ready to add modes to.
func NewModeParser() (p *ModeParser) {
	p = new(ModeParser)
	p.simple = make(map[int]string)
	p.parametered = make(map[int]string)
	p.list = make(map[int]string)
	p.nameToSimple = make(map[string]int)
	p.nameToParametered = make(map[string]int)
	p.nameToList = make(map[string]int)
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

// ParseModeLine parses a line of mode changes into core.UserDataChange structs.
// Redundant changes are compressed down into one.
// An error is returned if unknown modes are encountered, or modes are dropped
// due to missing parameters. The remainder of the modes are still parsed.
func (p *ModeParser) ParseModeLine(modeline []byte, params [][]byte) (*core.UserDataChange, os.Error) {
	var adding bool = true
	var unknown string
	var missing string
	var param int
	modes := string(modeline)
	changes := make (map[string]*core.UserDataChange)

	for _, char := range modes {
		if char == '+' {
			adding = true
			continue
		} else if char == '-' {
			adding = false
			continue
		}
		
		if v, ok := p.simple[char]; ok {
			change := new(core.UserDataChange)
			change.Name = v
			if adding {
				change.Data = "on"
			}

			changes[v] = change	
			continue
		}

		if v, ok := p.parametered[char]; ok {
			var change *core.UserDataChange
			if adding {
				if param >= len(params) {
					missing += string(char)
					continue
				}
				change = new(core.UserDataChange)
				change.Name = v
				change.Data = string(params[param])
				param++
			} else {
				change = new(core.UserDataChange)
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
			change := new(core.UserDataChange)
			change.Name = v + " " + string(params[param])
			param++

			if adding {
				change.Data = "on"
			}

			changes[v] = change
			continue
		}

		unknown += string(char)
	}

	// Turn the modes into a list.
	var c *core.UserDataChange
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
func (p *ModeParser) ParseChanges(c *core.UserDataChange) (modeline string) {
	var addmodes string
	var remmodes string
	var addparams string
	var remparams string

	for it := c; it != nil; it = c.Next {
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

		if strings.IndexRune(it.Name, ' ') != -1 {
			if v, ok := p.nameToList[it.Name[0:strings.IndexRune(it.Name, ' ')]]; ok {
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
