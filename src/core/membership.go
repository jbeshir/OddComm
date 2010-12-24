package core

import "oddcomm/lib/trie"


type Membership struct {
	u     *User
	c     *Channel
	unext *Membership // Next membership entry for the user.
	uprev *Membership // Next membership entry for the user.
	cprev *Membership // Previous membership entry for the channel.
	cnext *Membership // Next membership entry for the channel.
	data  trie.StringTrie
}

// Channel returns the channel that this membership entry is for.
func (m *Membership) Channel() (ch *Channel) {
	return m.c
}

// User returns the channel that this membership entry is for.
func (m *Membership) User() (u *User) {
	return m.u
}

// ChanNext gets the next membership entry for the channel's list.
func (m *Membership) ChanNext() (next *Membership) {
	return m.cnext
}

// UserNext gets the next membership entry for the user's list.
func (m *Membership) UserNext() (next *Membership) {
	return m.unext
}

// SetData sets the given single piece of metadata on the membership entry.
// source may be nil, in which case the metadata is set by the server.
// Setting it to "" unsets it.
func (m *Membership) SetData(source *User, name string, value string) {
	var oldvalue string

	if value != "" {
		oldvalue = m.data.Insert(name, value)
	} else {
		oldvalue = m.data.Remove(name)
	}

	// If nothing changed, don't call hooks.
	if oldvalue == value {
		return
	}

	c := new(DataChange)
	c.Name = name
	c.Data = value
	c.Member = m
	old := new(OldData)
	old.Data = value
	runMemberDataChangeHooks(m.c.Type(), source, m, name, oldvalue, value)
	runChanDataChangesHooks(m.c.Type(), source, m.c, c, old)
}

// SetDataList performs the given list of metadata changes on the membership
// entry. This is equivalent to lots of SetData calls, except hooks for all
// data changes will receive it as a single list, and it is cheaper.
// source may be nil, in which case the metadata is set by the server.
func (m *Membership) SetDataList(source *User, c *DataChange) {
	var oldvalues *OldData
	var lasthook *DataChange
	var lastold **OldData = &oldvalues

	for it := c; it != nil; it = it.Next {

		// Make the change.
		var oldvalue string
		if it.Data != "" {
			oldvalue = m.data.Insert(it.Name, it.Data)
		} else {
			oldvalue = m.data.Remove(it.Name)
		}

		// If this was a do-nothing change, cut it out.
		if oldvalue == it.Data {
			if lasthook != nil {
				lasthook.Next = it.Next
			} else {
				c = it.Next
			}
			continue
		}

		olddata := new(OldData)
		olddata.Data = oldvalue
		*lastold = olddata
		lasthook = it
		lastold = &olddata.Next
	}

	for it, old := c, oldvalues; it != nil && old != nil; it, old = it.Next, old.Next {
		runMemberDataChangeHooks(m.c.Type(), source, m, it.Name, old.Data, it.Data)
	}
	runChanDataChangesHooks(m.c.Type(), source, m.c, c, oldvalues)
}

// Data retrieves the requested piece of metadata from this membership entry.
// It returns "" if no such piece of metadata exists.
func (m *Membership) Data(name string) (value string) {
	return m.data.Get(name)
}

// DataRnge calls the given function for every piece of metadata with the
// given prefix. If none are found, the function is never called. Metadata
// items added while this function is running may or may not be missed.
func (m *Membership) DataRange(prefix string, f func(name, value string)) {
	for it := m.data.GetSub(prefix); it != nil; {
		name, data := it.Value()
		f(name, data)
		if !it.Next() {
			it = nil
		}
	}
}

// Remove removes this membership entry; the user is removed from the channel.
// The specified source is responsible. It may be nil.
func (m *Membership) Remove(source *User, message string) {

	m.c.mutex.Lock()
	m.u.mutex.Lock()

	if m.cprev == nil {
		m.c.users = m.cnext
	} else {
		m.cprev.cnext = m.cnext
	}
	if m.cnext != nil {
		m.cnext.cprev = m.cprev
	}

	if m.uprev == nil {
		m.u.chans = m.unext
	} else {
		m.uprev.unext = m.unext
	}
	if m.unext != nil {
		m.unext.uprev = m.uprev
	}

	m.u.mutex.Unlock()
	m.c.mutex.Unlock()

	runChanUserRemoveHooks(m.c.t, source, m.u, m.c, message)

	return
}
