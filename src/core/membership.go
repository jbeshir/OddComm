package core


type Membership struct {
	u *User
	c *Channel
	unext *Membership // Next membership entry for the user.
	uprev *Membership // Next membership entry for the user.
	cprev *Membership // Previous membership entry for the channel.
	cnext *Membership // Next membership entry for the channel.
	data map[string]string
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
