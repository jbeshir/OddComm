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
	// It isn't possible to change this once set, so we have no reason to
	// bother the core goroutine about this.
	return m.c
}

// User returns the channel that this membership entry is for.
func (m *Membership) User() (u *User) {
	// It isn't possible to change this once set, so we have no reason to
	// bother the core goroutine about this.
	return m.u
}

// ChanNext gets the next membership entry for the channel's list.
func (m *Membership) ChanNext() (next *Membership) {
	m.c.mutex.Lock()
	next = m.cnext
	m.c.mutex.Unlock()

	return
}

// UserNext gets the next membership entry for the user's list.
func (m *Membership) UserNext() (next *Membership) {
	m.u.mutex.Lock()
	next = m.unext
	m.u.mutex.Unlock()

	return
}

// SetData sets the given single piece of metadata on the membership entry.
// source may be nil, in which case the metadata is set by the server.
// Setting it to "" unsets it.
func (m *Membership) SetData(source *User, name string, value string) {
	var oldvalue string

	m.c.mutex.Lock()
	m.u.mutex.Lock()

	if value != "" {
		oldvalue = m.data.Add(name, value)
	} else {
		oldvalue = m.data.Del(name)
	}

	m.u.mutex.Unlock()
	m.c.mutex.Unlock()

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

	m.c.mutex.Lock()
	m.u.mutex.Lock()
	for it := c; it != nil; it = it.Next {

		// Make the change.
		var oldvalue string
		if it.Data != "" {
			oldvalue = m.data.Add(it.Name, it.Data)
		} else {
			oldvalue = m.data.Del(it.Name)
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
	m.u.mutex.Unlock()
	m.c.mutex.Unlock()

	for it, old := c, oldvalues; it != nil && old != nil; it, old = it.Next, old.Next {
		runMemberDataChangeHooks(m.c.Type(), source, m, it.Name, old.Data, it.Data)
	}
	runChanDataChangesHooks(m.c.Type(), source, m.c, c, oldvalues)
}

// Data retrieves the requested piece of metadata from this membership entry.
// It returns "" if no such piece of metadata exists.
func (m *Membership) Data(name string) (value string) {
	m.c.mutex.Lock()
	m.u.mutex.Lock()
	value = m.data.Get(name)
	m.u.mutex.Unlock()
	m.c.mutex.Unlock()

	return
}

// DataRange calls the given function for every piece of metadata with the
// given prefix. If none are found, the function is never called. Metadata
// items added while this function is running may or may not be missed.
func (m *Membership) DataRange(prefix string, f func(name, value string)) {
	var dataArray [100]DataChange
	var data []DataChange = dataArray[0:0]
	var root, it trie.StringTrie

	for firstrun := true; firstrun || !it.Empty(); {

		m.c.mutex.Lock()
		m.u.mutex.Lock()

		// On firstrun, get an iterator pointing to our first value.
		if firstrun {
			root = m.data.GetSub(prefix)
			it = root
			if !it.Empty() {
				if key, _ := it.Value(); key == "" {
					it = it.Next(root)
				}
			}
			firstrun = false
		}

		// Get up to 100 values from this subtrie.
		for i := 0; !it.Empty() && i < cap(data); i++ {
			data = data[0 : i+1]
			data[i].Name, data[i].Data = it.Value()
			it = it.Next(root)
		}

		m.u.mutex.Unlock()
		m.c.mutex.Unlock()

		// Call the function for all of them, and clear data.
		for _, item := range data {
			f(item.Name, item.Data)
		}
		data = data[0:0]
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
