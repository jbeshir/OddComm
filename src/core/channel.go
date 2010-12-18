package core

import "oddcomm/lib/trie"

import "strconv"
import "strings"
import "sync"
import "time"


// Represents a channel.
type Channel struct {
	mutex sync.Mutex
	name  string
	t     string
	ts    int64
	users *Membership
	data  trie.StringTrie
}


// GetChannel returns a channel with the given name and type. Type may be ""
// (for default). If the channel did not previously exist, it is created. If it
// already existed, it is simply returned.
func GetChannel(t, name string) (ch *Channel) {
	NAME := strings.ToUpper(name)

	chanMutex.Lock()
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
	chanMutex.Unlock()

	return
}

// FindChannel finds a channel with the given name and type, which may be ""
// for the default type. If none exist, it returns nil.
func FindChannel(t, name string) (ch *Channel) {
	NAME := strings.ToUpper(name)

	chanMutex.Lock()
	if _, ok := channels[t]; ok {
		ch = channels[t][NAME]
	}
	chanMutex.Unlock()

	return
}


// Name returns the channel's name.
func (ch *Channel) Name() string {
	// This cannot change after the channel has been created.
	// No need to bother with synchronisation.
	return ch.name
}

// Type returns the channel's type. This may be "", for default.
func (ch *Channel) Type() string {
	// This cannot change after the channel has been created.
	// No need to bother with synchronisation.
	return ch.t
}

// TS returns the channel's creation time.
func (ch *Channel) TS() (ts int64) {
	// This actually CAN change after the channel has been created,
	// if the channel is merged with a channel created earlier.
	// Thus, synchronisation.
	ch.mutex.Lock()
	ts = ch.ts
	ch.mutex.Unlock()

	return
}

// SetData sets the given single piece of metadata on the channel.
// source may be nil, in which case the metadata is set by the server.
// Setting it to "" unsets it.
func (ch *Channel) SetData(source *User, name string, value string) {
	var oldvalue string

	ch.mutex.Lock()
	if value != "" {
		oldvalue = ch.data.Add(name, value)
	} else {
		oldvalue = ch.data.Del(name)
	}
	ch.mutex.Unlock()

	// If nothing changed, don't call hooks.
	if oldvalue == value {
		return
	}

	c := new(DataChange)
	c.Name = name
	c.Data = value
	old := new(OldData)
	old.Data = value
	runChanDataChangeHooks(ch.Type(), source, ch, name, oldvalue, value)
	runChanDataChangesHooks(ch.Type(), source, ch, c, old)
}

// SetDataList performs the given list of metadata changes on the channel.
// This is equivalent to lots of SetData calls, except hooks for all data
// changes will receive it as a single list, and it is cheaper.
// source may be nil, in which case the metadata is set by the server.
func (ch *Channel) SetDataList(source *User, c *DataChange) {
	var oldvalues *OldData
	var lasthook *DataChange
	var lastold **OldData = &oldvalues

	ch.mutex.Lock()
	for it := c; it != nil; it = it.Next {

		// Figure out what we're making the change to.
		// The channel, or a member?
		// If a member, lock them.
		t := ch.data
		if it.Member != nil {
			t = it.Member.data
			it.Member.u.mutex.Lock()
		}

		// Make the change.
		var oldvalue string
		if it.Data != "" {
			oldvalue = t.Add(it.Name, it.Data)
		} else {
			oldvalue = t.Del(it.Name)
		}

		// If we locked a member, unlock them.
		if it.Member != nil {
			it.Member.u.mutex.Unlock()
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
		*lastold = olddata
		lasthook = it
		lastold = &olddata.Next
	}
	ch.mutex.Unlock()

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
	ch.mutex.Lock()
	value = ch.data.Get(name)
	ch.mutex.Unlock()

	return
}

// DataRange calls the given function for every piece of metadata with the
// given prefix. If none are found, the function is never called. Metadata
// items added while this function is running may or may not be missed.
func (ch *Channel) DataRange(prefix string, f func(name, value string)) {
	var dataArray [100]DataChange
	var data []DataChange = dataArray[0:0]
	var root, it trie.StringTrie

	for firstrun := true; firstrun || !it.Empty(); {

		ch.mutex.Lock()

		if firstrun {
			root = ch.data.GetSub(prefix)
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

		ch.mutex.Unlock()

		// Call the function for all of them, and clear data.
		for _, item := range data {
			f(item.Name, item.Data)
		}
		data = data[0:0]
	}
}

// Users returns a pointer to the channel's membership list.
func (ch *Channel) Users() (users *Membership) {
	ch.mutex.Lock()
	users = ch.users
	ch.mutex.Unlock()

	return
}

// GetMember returns a pointer to this channel's membership structure for the
// given user, or nil if they are not a member. This is also how to check
// whether a user is on the channel or not.
func (ch *Channel) GetMember(u *User) (m *Membership) {
	ch.mutex.Lock()
	for it := ch.users; it != nil; it = it.cnext {
		if it.u == u {
			m = it
			break
		}
	}
	ch.mutex.Unlock()

	return
}

// Join adds a user to the channel.
func (ch *Channel) Join(u *User) {
	var alreadyJoined bool
	var joined bool

	ch.mutex.Lock()
	u.mutex.Lock()

	// Unregistered users may not join channels.
	if u.regstate == 0 {

		// Users who are already IN the channel may not join.
		for it := ch.users; it != nil; it = it.cnext {
			if it.u == u {
				alreadyJoined = true
			}
		}
	}

	if !alreadyJoined {
		m := new(Membership)
		m.c = ch
		m.u = u
		m.unext = u.chans
		m.cnext = ch.users
		u.chans = m
		ch.users = m
		if m.unext != nil {
			m.unext.uprev = m
		}

		u.mutex.Unlock()

		if m.cnext != nil {
			m.cnext.u.mutex.Lock()
			m.cnext.cprev = m
			m.cnext.u.mutex.Unlock()
		}
		joined = true
	} else {
		u.mutex.Unlock()
	}

	ch.mutex.Unlock()


	if joined {
		runChanUserJoinHooks(ch.t, u, ch)
	}
}

// Remove removes the given user from the channel.
// source may be nil, indicating that they are being removed by the server.
// This behaves as iterates the user list, and then calling Remove() on the
// Membership struct would, but is faster.
func (ch *Channel) Remove(source, u *User, message string) {

	// Unregistered users may not remove other users.
	if source != nil && !source.Registered() {
		return
	}

	ch.mutex.Lock()
	u.mutex.Lock()

	// Unregistered users may not join channels in the first place.
	if !u.Registered() {
		return
	}

	// Search for them, remove them if we find them.
	for it := ch.users; it != nil; it = it.cnext {
		if it.u == u {
			// FIXME FIXME FIXME BUG
			// This must be replaced by the contents of it.Remove()
			// As we are holding mutexes it will attempt to lock.
			// This call WILL lock up.
			it.Remove(source, message)
		}
	}

	u.mutex.Unlock()
	ch.mutex.Unlock()
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
func (ch *Channel) GetTopic() (topic, setby, setat string) {
	ch.mutex.Lock()
	topic = ch.data.Get("topic")
	setby = ch.data.Get("topic setby")
	setat = ch.data.Get("topic setat")
	ch.mutex.Unlock()

	if setby == "" {
		setby = "Server.name"
	}
	if setat == "" {
		setat = "0"
	}
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
