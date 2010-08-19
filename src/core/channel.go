package core

import "strconv"
import "strings"
import "time"


// Represents a channel.
type Channel struct {
	name string
	t string
	ts int64
	users *Membership
	data *Trie
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

// Type returns the channel's type. This may be "", for default.
func (ch *Channel) Type() (t string) {
	// This cannot change after the channel has been created.
	// No need to bother the core goroutine with synchronisation.
	return ch.t
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

		var old interface{}
		if value != "" {
			old = TrieAdd(&ch.data, name, value)
		} else {
			old = TrieDel(&ch.data, name)
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

	runChanDataChangeHooks(ch.Type(), source, ch, name, oldvalue, value)

	c := new(DataChange)
	c.Name = name
	c.Data = value
	old := new(OldData)
	old.Data = value
	runChanDataChangesHooks(ch.Type(), source, ch, c, old)
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

			// Figure out what we're making the change to.
			// The channel, or a member?
			trie := &ch.data
			if it.Member != nil {
				trie = &it.Member.data
			}

			// Make the change.
			var old interface{}
			var oldvalue string
			if it.Data != "" {
				old = TrieAdd(trie, it.Name, it.Data)
			} else {
				old = TrieDel(trie, it.Name)
			}
			if old != nil {
				oldvalue = old.(string)
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

			// Otherwise, add the old value to the old data list.
			olddata := new(OldData)
			olddata.Data = oldvalue
			olddata.Next = oldvalues
			oldvalues = olddata
			lasthook = it
		}

		wait <- true
	}
	<-wait

	for it, old := c, oldvalues; it != nil && old != nil; it, old = it.Next, old.Next {
		if it.Member == nil {
			runChanDataChangeHooks(ch.Type(), source, ch, it.Name,
			                       old.Data, it.Data)
		} else {
			runMemberDataChangeHooks(ch.Type(), source, it.Member,
                                                 it.Name, old.Data, it.Data)
		}
	}
	runChanDataChangesHooks(ch.Type(), source, ch, c, oldvalues)
}

// Data gets the given piece of metadata.
// If it is not set, this method returns "".
func (ch *Channel) Data(name string) (value string) {
	wait := make(chan bool)
	corechan <- func() {
		val := TrieGet(&ch.data, name)
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
func (ch *Channel) DataRange(prefix string, f func(name, value string)) {
	var dataArray [50]DataChange
	var data []DataChange = dataArray[0:0]
	var root, it *Trie
	wait := make(chan bool)

	// Get an iterator pointing to our first value.
	corechan <- func() {
		root = TrieGetSub(&ch.data, prefix)
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

// GetMember returns a pointer to this channel's membership structure for the
// given user, or nil if they are not a member. This is also how to check
// whether a user is on the channel or not.
func (ch *Channel) GetMember(u *User) (m *Membership) {
	wait := make(chan bool)
	corechan <- func() {
		for it := ch.users; it != nil; it = it.cnext {
			if it.u == u {
				m = it
				break
			}
		}
		wait <- true
	}
	<-wait
	return
}

// Join adds a user to the channel.
func (ch *Channel) Join(u *User) {
	var joined bool
	wait := make(chan bool)
	corechan <- func() {

		// Unregistered users may not join channels.
		if u.regcount != 0 {
			wait <- true
			return
		}

		// Users who are already IN the channel may not join.
		for it := ch.users; it != nil; it = it.cnext {
			if it.u == u {
				wait <- true
				return
			}
		}

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
		joined = true
		wait <- true
	}
	<-wait

	if joined {
		runChanUserJoinHooks(ch.t, u, ch)
	}
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
	runChanMessageHooks(ch.t, t, source, ch, message)
}

// Delete deletes the channel.
// Does nothing if users are still in the channel.
func (ch *Channel) Delete() {
	// This doesn't actually do anything yet.
}

// GetTopic gets the topic, the topic setter string, and the time it was set.
func (ch *Channel) GetTopic() (topic, setby, setat string){
	wait := make(chan bool)
	corechan <- func() {
		setby = "Server.name"
		setat = "0"
		if v := TrieGet(&ch.data, "topic"); v != nil {
			topic = v.(string)
		}
		if v := TrieGet(&ch.data, "topic setby"); v != nil {
			setby = v.(string)
		}
		if v := TrieGet(&ch.data, "topic setat"); v != nil {
			setat = v.(string)
		}
		wait <- true
	}
	<-wait
	return
}

// SetTopic sets the topic, including recording its setting and set time.
func (ch *Channel) SetTopic(source *User, topic string) {
	var changes [3]DataChange
	changes[0].Name = "topic"
	changes[0].Data = topic
	changes[0].Next = &changes[1]
	changes[1].Name = "topic setat"
	changes[1].Data = strconv.Itoa64(time.Seconds())
	if source != nil {
		changes[1].Next = &changes[2]
		changes[2].Name = "topic setby"
		changes[2].Data = source.GetSetBy()
	}
	ch.SetDataList(source, &changes[0])
}
