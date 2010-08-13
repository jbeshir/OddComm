package client

/* Handle extbans. */
extensions map[int]map[int]string
extToChar map[int]map[string]int

func init() {
	extensions = make(map[int]map[int]string)
	extToChar = make(map[int]map[string]int)
}



// Parse list and membership modes for extended modes, returning the value of
// the metadata to set/unset, and the 'rest' of the string, which indicates
// either the end of the list mode's name, or the user the membership mode is
// to apply to. If the mode is not extended, the value is as default and the
// "rest" is the whole param.
func (p *ModeParser) parseExtModes(mode int, par string) (data, rest string) {
	if _, ok := p.extensions[mode]; !ok {
		rest = par
		data = "on"
		return
	}

	rest = par[strings.IndexRune(par, ':') + 1:]
	extmodes := par[0:strings.IndexRune(par, ':') + 1]

	// The extmodes string will include the colon. This permits us to
	// distinguish lack of a list from an empty one.
	if len(extmodes) > 0 {
		for ext := range extmodes[0:len(extmodes)-1] {
			if val, ok := p.extensions[mode][ext]; ok {
				if data != "" {
					data += " "
				}
				data += val
			}
		}
	} else {
		if val, ok := p.extensions[mode][0]; ok {
			data = val
		} else {
			data = "on"
		}
	}

	return
} 
