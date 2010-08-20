package core


type Membership struct {
	u *User
	c *Channel
	unext *Membership // Next membership entry for the user.
	uprev *Membership // Next membership entry for the user.
	cprev *Membership // Previous membership entry for the channel.
	cnext *Membership // Next membership entry for the channel.
	data *Trie
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
	wait := make(chan bool)
	corechan <- func() {
		next = m.cnext
		wait <- true
	}
	<-wait
	return
}

// UserNext gets the next membership entry for the user's list.
func (m *Membership) UserNext() (next *Membership) {
	wait := make(chan bool)
	corechan <- func() {
		next = m.unext
		wait <- true
	}
	<-wait
	return
}

// SetData sets the given single piece of metadata on the membership entry.
// source may be nil, in which case the metadata is set by the server.
// Setting it to "" unsets it.
func (m *Membership) SetData(source *User, name string, value string) {
	var oldvalue string

	wait := make(chan bool)
	corechan <- func() {

		var old interface{}
		if value != "" {
			old = TrieAdd(&m.data, name, value)
		} else {
			old = TrieDel(&m.data, name)
		}
		if old != nil {
			oldvalue = old.(string)
		}

		wait <- true
	}
	<-wait

	// If nothing changed, don't call hooks.
	if oldvalue == value {
		return
	}

	runMemberDataChangeHooks(m.c.Type(), source, m, name, oldvalue, value)

	c := new(DataChange)
	c.Name = name
	c.Data = value
	c.Member = m
	old := new(OldData)
	old.Data = value
	runChanDataChangesHooks(m.c.Type(), source, m.c, c, old)
}


// Data retrieves the requested piece of metadata from this membership entry.
// It returns "" if no such piece of metadata exists.
func (m *Membership) Data(name string) (value string) {
	wait := make(chan bool)
	corechan <- func() {
		val := TrieGet(&m.data, name)
		if val != nil {
			value = val.(string)
		}
		wait <- true
	}
	<-wait
	return

}

// DataRange calls the given function for every piece of metadata with the
// given prefix. If none are found, the function is never called. Metadata
// items added while this function is running may or may not be missed.
func (m *Membership) DataRange(prefix string, f func(name, value string)) {
	var dataArray [50]DataChange
	var data []DataChange = dataArray[0:0]
	var root, it *Trie
	wait := make(chan bool)

	// Get an iterator pointing to our first value.
	corechan <- func() {
		root = TrieGetSub(&m.data, prefix)
		it = root
		if it != nil {
			if key, _ := it.Value(); key == "" {
				it = it.Next(root)
			}
		}
		wait <- true
	}
	<-wait

	for it != nil {
		// Get up to 50 values from this subtrie.
		corechan <- func() {
			for i := 0; i < cap(data); i++ {
				var val interface{}
				data = data[0:i+1]
				data[i].Name, val = it.Value()
				data[i].Data = val.(string)
				it = it.Next(root)
				if it == nil {
					break
				}
			}
			wait <- true
		}
		<- wait

		// Call the function for all of them, and clear data.
		for _, item := range data {
			f(item.Name, item.Data)
		}
		data = data[0:0]
	}
}

// Remove removes this membership entry; the user is removed from the channel.
// The specified source is responsible. It may be nil.
func (m *Membership) Remove(source *User) {
	wait := make(chan bool)
	corechan <- func() {
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

		wait <- true
	}
	<-wait

	runChanUserRemoveHooks(m.c.t, source, m.u, m.c)

	return
}
