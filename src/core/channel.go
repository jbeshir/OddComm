package core

import "oddcomm/lib/trie"

import "strconv"
import "strings"
import "sync"
import "time"
import "unsafe"


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
	ch = (*Channel)(channels.Get(t + " " + NAME))
	if ch == nil {
		ch = new(Channel)
		ch.name = name
		ch.t = t
		ch.ts = time.Seconds()
		channels.Insert(t + " " + NAME, unsafe.Pointer(ch))
	}
	chanMutex.Unlock()

	return
}

// FindChannel finds a channel with the given name and type, which may be ""
// for the default type. If none exist, it returns nil.
func FindChannel(t, name string) (ch *Channel) {
	NAME := strings.ToUpper(name)
	return (*Channel)(channels.Get(t + " " + NAME))
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
	return ch.ts
}

// SetData sets the given single piece of metadata on the channel.
// source may be nil, in which case the metadata is set by the server.
// Setting it to "" unsets it.
func (ch *Channel) SetData(origin interface{}, source *User, name, value string) {
	var oldvalue string

	ch.mutex.Lock()

	if value != "" {
		oldvalue = ch.data.Insert(name, value)
	} else {
		oldvalue = ch.data.Remove(name)
	}

	// If nothing changed, don't call hooks.
	if oldvalue == value {
		return
	}

	var c DataChange
	c.Name = name
	c.Data = value

	hookRunner <- func() {
		runChanDataChangeHooks(origin, ch.Type(), source, ch, name,
			oldvalue, value)
		runChanDataChangesHooks(origin, ch.Type(), source, ch,
			[]DataChange{c}, []string{oldvalue})
	}

	ch.mutex.Unlock()
}

// SetDataList performs the given list of metadata changes on the channel.
// This is equivalent to lots of SetData calls, except hooks for all data
// changes will receive it as a single list, and it is cheaper.
// source may be nil, in which case the metadata is set by the server.
func (ch *Channel) SetDataList(origin interface{}, source *User, changes []DataChange) {
	done := make([]DataChange, 0, len(changes))
	old := make([]string, 0, len(changes))

	ch.mutex.Lock()

	for _, it := range changes {

		// Figure out what we're making the change to.
		// The channel, or a member?
		// If a member, lock them.
		t := &ch.data
		if it.Member != nil {
			t = &it.Member.data
			it.Member.u.mutex.Lock()
		}

		// Make the change.
		var oldvalue string
		if it.Data != "" {
			oldvalue = (*t).Insert(it.Name, it.Data)
		} else {
			oldvalue = (*t).Remove(it.Name)
		}

		// If we locked a member, unlock them.
		if it.Member != nil {
			it.Member.u.mutex.Unlock()
		}

		// If this was a do-nothing change, don't report it.
		if oldvalue == it.Data {
			continue
		}

		// Otherwise, add to the slices.
		done = append(done, it)
		old = append(old, oldvalue)
	}

	hookRunner <- func() {
		for i, it := range done {
			if it.Member == nil {
				runChanDataChangeHooks(origin, ch.Type(), source,
					ch, it.Name, old[i], it.Data)
			} else {
				runMemberDataChangeHooks(origin, ch.Type(), source,
					it.Member, it.Name, old[i], it.Data)
			}
		}
		runChanDataChangesHooks(origin, ch.Type(), source, ch, done, old)
	}

	ch.mutex.Unlock()
}

// Data gets the given piece of metadata.
// If it is not set, this method returns "".
func (ch *Channel) Data(name string) (value string) {
	return ch.data.Get(name)
}

// DataRange calls the given function for every piece of metadata with the
// given prefix. If none are found, the function is never called. Metadata
// items added while this function is running may or may not be missed.
func (ch *Channel) DataRange(prefix string, f func(name, value string)) {
	for it := ch.data.IterSub(prefix); it != nil; {
		name, data := it.Value()
		f(name, data)
		if !it.Next() {
			it = nil
		}
	}
}

// Users returns a pointer to the channel's membership list.
func (ch *Channel) Users() (users *Membership) {
	return ch.users
}

// GetMember returns a pointer to this channel's membership structure for the
// given user, or nil if they are not a member. This is also how to check
// whether a user is on the channel or not.
func (ch *Channel) GetMember(u *User) (m *Membership) {

	// We actually iterate the user's channel list, because it's often
	// shorter, and much more reasonable to put a low limit on.
	for it := u.chans; it != nil; it = it.unext {
		if it.c == ch {
			m = it
			break
		}
	}
	return
}

// Join adds a set of users to the channel.
// Multiple joins by the same user are not guaranteed to be reported
// in the order they happened.
// Returns the users that actually joined.
func (ch *Channel) Join(origin interface{}, users []*User) []*User {
	var joinedusers []*User

	ch.mutex.Lock()

	for _, u := range users {
		u.mutex.Lock()

		// Unregistered users may not join channels.
		if !u.Registered() {
			u.mutex.Unlock()
			continue
		}

		// Users who are already IN the channel may not join.
		if ch.GetMember(u) != nil {
			u.mutex.Unlock()
			continue
		}

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

		joinedusers = append(joinedusers, u)
	}

	hookRunner <- func() {
		runChanUserJoinHooks(origin, ch.t, ch, joinedusers)
	}

	ch.mutex.Unlock()

	return joinedusers
}

// Remove removes the given user from the channel.
// source may be nil, indicating that they are being removed by the server.
// This behaves as iterates the user list, and then calling Remove() on the
// Membership struct would, but is faster.
func (ch *Channel) Remove(origin interface{}, source, u *User, message string) {
	var m *Membership

	// Unregistered users may not remove other users.
	if source != nil && !source.Registered() {
		return
	}

	ch.mutex.Lock()
	u.mutex.Lock()

	// Unregistered users may not join channels in the first place.
	// This also prevents deleted users from being removed.
	if u.Registered() {

		// Search for them, remove them if we find them.
		if m = ch.GetMember(u); m != nil {
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
		}
	}

	if m != nil {
		hookRunner <- func() {
			runChanUserRemoveHooks(origin, m.c.t, source, m.u, m.c,
				message)
		}
	}

	u.mutex.Unlock()
	ch.mutex.Unlock()
}

// Message sends a message to the channel.
// source may be nil, indicating a message from the server.
// t may be "" (for default), and indicates the type of message.
func (ch *Channel) Message(origin interface{}, source *User, message []byte, t string) {

	// Unregistered users may not send messages.
	if !source.Registered() {
		return
	}

	// We actually just call hooks, and let the subsystems handle it.
	runChanMessageHooks(origin, ch.t, t, source, ch, message)
}

// Delete deletes the channel.
// Does nothing if users are still in the channel.
func (ch *Channel) Delete() {
	// This doesn't actually do anything yet.
}

// GetTopic gets the topic, the topic setter string, and the time it was set.
func (ch *Channel) GetTopic() (topic, setby, setat string) {
	topic = ch.data.Get("topic")
	setby = ch.data.Get("topic setby")
	setat = ch.data.Get("topic setat")

	if setby == "" {
		setby = "Server.name"
	}
	if setat == "" {
		setat = "0"
	}
	return
}

// SetTopic sets the topic, including recording its setting and set time.
func (ch *Channel) SetTopic(origin interface{}, source *User, topic string) {
	changes := make([]DataChange, 3)
	changes[0].Name = "topic"
	changes[0].Data = topic
	changes[1].Name = "topic setat"
	changes[1].Data = strconv.Itoa64(time.Seconds())
	changes[2].Name = "topic setby"
	if source != nil {
		changes[2].Data = source.GetSetBy()
	}
	ch.SetDataList(origin, source, changes)
}
