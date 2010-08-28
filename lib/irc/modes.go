package irc

import "strings"

import "oddcomm/src/core"


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
	p.clearPair(mode, metadata)
	p.simple[mode] = metadata
	p.nameToSimple[metadata] = mode
}

// AddParametered adds a parametered mode. It cannot be called concurrently
// with itself, or any lookups on the parser.
// If a mode character or metadata name is added twice, the mode associated
// with the previous value is deleted.
func (p *ModeParser) AddParametered(mode int, metadata string) {
	p.clearPair(mode, metadata)
	p.parametered[mode] = metadata
	p.nameToParametered[metadata] = mode
}

// AddList adds a list mode. It cannot be called concurrently with itself, or
// any lookups on the parser.
// If a mode character or metadata name is added twice, the mode associated
// with the previous value is deleted.
func (p *ModeParser) AddList(mode int, metadata string) {
	p.clearPair(mode, metadata)
	p.list[mode] = metadata
	p.nameToList[metadata] = mode
}

// AddMembership adds a membership mode. It cannot be called concurrently with
// itself, or any lookups on the parser.
// If a mode character or metadata name is added twice, the mode associated
// with the previous value is deleted.
// Membership modes are ignored if this is not a channel mode parser.
func (p *ModeParser) AddMembership(mode int, metadata string) {
	p.clearPair(mode, metadata)
	p.membership[mode] = metadata
	p.nameToMembership[metadata] = mode
}

// ExtendModeToData extends the conversion of changes to the given mode to
// metadata. The modeToName function should expect a boolean indicating whether
// the mode is being added or removed, and the user/channel being changed
// followed by the mode's parameter. It must return a list of DataChange
// objects to apply, which may be empty.
//
// XXX: Doesn't work well enough for multiple extban/opflag changes in a line.
// API changes under consideration.
func (p *ModeParser) ExtendModeToData(mode int, modeToName func(bool, core.Extensible, string) (*core.DataChange)) {
	p.extended[mode] = modeToName
}

// ExtendDataToMode extends the conversion of the given data into a mode.
// The nameToMode function shuld expect the user/channel being changed
// followed by the metadata item's full name, its previous value, and its
// current value. It should return a slice of added mode characters and
// another of removed mode characters, and a slice of added mode parameters
// and another of removed mode parameters. Any of these may be empty.
//
// The added or removed mode characters may include modes other than the mode
// which corresponds to this name if and only if the previous value is not "".
//
// For list modes, this function will also be called to convert existing
// metadata into list mode entries, with a previous value of "", which is the
// reason for the above restriction.
func (p *ModeParser) ExtendDataToMode(name string, nameToMode func(core.Extensible, string, string, string) ([]int, []string, []int, []string)) {
	p.nameToExt[name] = nameToMode
}

// ExtendGetSet extends checks on the current state of a simple or parametered
// mode. The getSet function should expect a user or channel, and return a
// value for the mode, which may be "" to indicate it is unset, or any other
// value to indicate it is set, and for parametered modes, with that as the
// parameter.
//
// This method cannot be called concurrently with itself, or any lookups on
// the parser.
func (p *ModeParser) ExtendGetSet(mode int, getSet func(core.Extensible) string) {
	p.getExt[mode] = getSet
}


// GetName returns the base name for the metadata corresponding to the given
// mode, permitting it to be used for permission checks and messages.
// As converting modes (and parameters) into metadata is complex, this should
// not be used in place of ParseModeLine.
// If the mode does not exist, returns "".
func (p *ModeParser) GetName(mode int) string {
	if v, ok := p.simple[mode]; ok { return v }
	if v, ok := p.parametered[mode]; ok { return v }
	if v, ok := p.list[mode]; ok { return v }
	if v, ok := p.membership[mode]; ok { return v }
	return ""
}

// GetMode returns the mode letter corresponding to the given metadata's name,
// permitting it to be used for permission checks and messages.
// As converting metadata into modes (and parameters) is complex, this should
// not be used in place of ParseChanges.
// If the mode does not exist, returns 0.
func (p *ModeParser) GetMode(name string) int {
	for {
		if v, ok := p.nameToSimple[name]; ok { return v }
		if v, ok := p.nameToParametered[name]; ok { return v }
		if v, ok := p.nameToList[name]; ok { return v }
		if v, ok := p.nameToMembership[name]; ok { return v }

		// Keep searching for modes matching a space-separated prefix.
		if v := strings.LastIndex(name, " "); v != -1 {
			name = name[0:v]
		} else {
			break
		}
	}
	return 0
}


// Clears all references to the given mode pair. Used when a mode is added, to
// delete any remaining bits of a previous added mode.
func (p *ModeParser) clearPair(mode int, name string) {

	if v, ok := p.simple[mode]; ok { p.clearName(v) }
	if v, ok := p.parametered[mode]; ok { p.clearName(v) }
	if v, ok := p.list[mode]; ok { p.clearName(v) }
	if v, ok := p.membership[mode]; ok { p.clearName(v) }

	if v, ok := p.nameToSimple[name]; ok { p.clearMode(v) }
	if v, ok := p.nameToParametered[name]; ok { p.clearMode(v) }
	if v, ok := p.nameToList[name]; ok { p.clearMode(v) }
	if v, ok := p.nameToMembership[name]; ok { p.clearMode(v) }

	p.clearMode(mode)
	p.clearName(name)
}

// Clears any reference to the given mode letter.
func (p *ModeParser) clearMode(mode int) {
	p.simple[mode] = "", false
	p.parametered[mode] = "", false
	p.list[mode] = "", false
	p.membership[mode] = "", false
	p.extended[mode] = nil, false
	p.getExt[mode] = nil, false
}

// Clears any reference to the given metadata name.
func (p *ModeParser) clearName(name string) {
	p.nameToSimple[name] = 0, false
	p.nameToParametered[name] = 0, false
	p.nameToList[name] = 0, false
	p.nameToMembership[name] = 0, false
	p.nameToExt[name] = nil, false
}
