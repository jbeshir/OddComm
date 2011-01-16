package irc

import "strconv"

import "oddcomm/src/core"


type prefix struct {
	prefix   int
	metadata string
	level    int
	next     *prefix
}

// AddPrefix adds a prefix to NAMES/WHO, associated with the given piece of
// metadata being set on the user's membership object. level sets the position
// of this prefix in comparison to other prefixes.
func (p *ModeParser) AddPrefix(prefixchar int, metadata string, level int) {

	// Make new prefix.
	pr := new(prefix)
	pr.prefix = prefixchar
	pr.metadata = metadata
	pr.level = level

	// Find the appropriate place to put the new prefix.
	place := &p.prefixes
	for *place != nil && (*place).level > pr.level {
		place = &(*place).next
	}

	// Add the new prefix.
	pr.next = *place
	*place = pr

	// Update the prefix when the metadata is set or unset.
	core.HookMemberDataChange("", metadata, func(_ string, u *core.User, m *core.Membership, _, _ string) {
		p.updatePrefix(m)
	})
}


// AllPrefixes returns a list of all the prefixes, and their corresponding
// mode. Prefixes whose corresponding metadata does not have a mode are
// skipped; these should not exist anyway. Prefixes will appear in order of
// rank, from highest to lowest.
func (p *ModeParser) AllPrefixes() (prefixes string, modes string) {
	for it := p.prefixes; it != nil; it = it.next {
		m, ok := p.nameToMembership[it.metadata]
		if !ok {
			continue
		}

		prefixes += string(it.prefix)
		modes += string(m)
	}
	return
}

// GetPrefixes returns the prefixes for a membership entry.
func (p *ModeParser) GetPrefixes(m *core.Membership) string {
	return m.Data(strconv.Uitoa64(p.id) + " prefixes")
}


// Update the prefix on a membership entry.
func (p *ModeParser) updatePrefix(m *core.Membership) {
	var prefix string
	for it := p.prefixes; it != nil; it = it.next {
		if m.Data(it.metadata) != "" {
			prefix += string(it.prefix)
		}
	}
	m.SetData("", nil, strconv.Uitoa64(p.id)+" prefixes", prefix)
}
