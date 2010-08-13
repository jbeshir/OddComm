package core

import "strings"
import "time"


// Represents a channel.
type Channel struct {
	name string
	t string
	ts int64
	users *Membership
	data map[string]string
}


// GetChannel returns a channel with the given name and type. Type may be ""
// (for default). If the channel did not previously exist, it is created. If it
// already existed, it is simply returned.
func GetChannel(t, name string) (ch *Channel) {
	NAME := strings.ToUpper(name)
	if _, ok := channels[t]; ok {
		if v, ok := channels[t][NAME]; ok {
			ch = v
		} else {
			ch = new(Channel)
			ch.name = name
			ch.t = t
			ch.ts = time.Seconds()
			channels[t][NAME] = ch
		}
	} else {
		channels[t] = make(map[string]*Channel)
		ch = new(Channel)
		ch.name = name
		ch.t = t
		ch.ts = time.Seconds()
		channels[t][NAME] = ch
	}
	return
}

// FindChannel finds a channel with the given name and type, which may be ""
// for the default type. If none exist, it returns nil.
func FindChannel(t, name string) (ch *Channel) {
	NAME := strings.ToUpper(name)
	if _, ok := channels[t]; ok {
		if v, ok := channels[t][NAME]; ok {
			ch = v
		}
	}
	return
}


// Name returns the channel's name.
func (ch *Channel) Name() (name string) {
	wait := make(chan bool)
	corechan <- func() {
		name = ch.name
		wait <- true
	}
	<-wait

	return
}

// TS returns the channel's creation time.
func (ch *Channel) TS() (ts int64) {
	wait := make(chan bool)
	corechan <- func() {
		ts = ch.ts
		wait <- true
	}
	<-wait
	return
}

// SetData sets the given single piece of metadata on the channel.
// source may be nil, in which case the metadata is set by the server.
// Setting it to "" unsets it.
func (ch *Channel) SetData(source *User, name string, value string) {
	var oldvalue string

	wait := make(chan bool)
	corechan <- func() {
		oldvalue = ch.data[name]
		if value != "" {
			ch.data[name] = value
		} else {
			ch.data[name] = "", false
		}

		wait <- true
	}
	<-wait

	// If nothing changed, don't call hooks.
	if oldvalue == value {
		return
	}

	// runChanDataChangeHooks(source, ch, name, oldvalue, value)

	c := new(DataChange)
	c.Name = name
	c.Data = value
	old := new(OldData)
	old.Data = value
	// runChanDataChangesHooks(source, ch, c, old)
}

// SetDataList performs the given list of metadata changes on the channel.
// This is equivalent to lots of SetData calls, except hooks for all data
// changes will receive it as a single list, and it is cheaper.
// source may be nil, in which case the metadata is set by the server.
func (ch *Channel) SetDataList(source *User, c *DataChange) {
	var oldvalues *OldData
	wait := make(chan bool)
	corechan <- func() {
		var lasthook *DataChange
		for it := c; it != nil; it = it.Next {

			// If this is a do-nothing change, cut it out.
			if ch.data[it.Name] == it.Data {
				if lasthook != nil {
					lasthook.Next = it.Next
				} else {
					c = it.Next
				}
			}

			old := new(OldData)
			old.Data = ch.data[it.Name]
			old.Next = oldvalues
			oldvalues = old

			if it.Data != "" {
				ch.data[it.Name] = it.Data
			} else {
				ch.data[it.Name] = "", false
			}

			lasthook = it
		}

		wait <- true
	}
	<-wait

	for it, old := c, oldvalues; it != nil && old != nil; it, old = it.Next, old.Next {
		// runUserChanChangeHooks(source, ch, c.Name, old.Data, c.Data)
	}
	// runChanDataChangesHooks(source, ch, c, oldvalues)
}

// Data gets the given piece of metadata.
// If it is not set, this method returns "".
func (ch *Channel) Data(name string) (value string) {
	wait := make(chan bool)
	corechan <- func() {
		value = ch.data[name]
		wait <- true
	}
	<-wait

	return
}

// Users returns a pointer to the channel's membership list.
func (ch* Channel) Users() (users *Membership) {
	wait := make(chan bool)
	corechan <- func() {
		users = ch.users
		wait <- true
	}
	<-wait
	return
}

// Join adds a user to the channel.
func (ch *Channel) Join(u *User) {

	// Unregistered users may not join channels.
	if !u.Registered() {
		return
	}

	wait := make(chan bool)
	corechan <- func() {
		m := new(Membership)
		m.c = ch
		m.u = u
		m.unext = u.chans
		m.cnext = ch.users
		u.chans = m
		ch.users = m
		if m.cnext != nil {
			m.cnext.cprev = m
		}
		if m.unext != nil {
			m.unext.uprev = m
		}
		wait <- true
	}
	<-wait

	runChanUserJoinHooks(u, ch)
}

// Remove removes the given user from the channel.
// source may be nil, indicating that they are being removed by the server.
// This iterates the user list, and then calls Remove() on the Membership
// struct, as a convienience function.
func (ch *Channel) Remove(source, u *User) {

	// Unregistered users may not join channels OR remove other users.
	if !source.Registered() || !u.Registered() {
		return
	}
	
	// Search for them, remove them if we find them.
	for it := ch.Users(); it != nil; it = it.ChanNext() {
		if it.User() == u {
			it.Remove(source)
		}
	}
}

// Message sends a message to the channel.
// source may be nil, indicating a message from the server.
// t may be "" (for default), and indicates the type of message.
func (ch *Channel) Message(source *User, message []byte, t string) {

	// Unregistered users may not send messages.
	if !source.Registered() {
		return
	}

	// We actually just call hooks, and let the subsystems handle it.
	runChanMessageHooks(source, ch, message, t)
}

// Delete deletes the channel.
// Remaining users are kicked, if any still exist.
func (ch *Channel) Delete() {
	// This doesn't actually do anything yet.
}
