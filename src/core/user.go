package core

import "os"
import "strings"
import "sync"

import "oddcomm/lib/trie"


// Represents a user.
type User struct {
	mutex    sync.Mutex
	id       string
	nick     string
	checked  bool
	regstate int
	chans    *Membership
	data     trie.StringTrie
}


// NewUser creates a new user, with creator the name of its creating package.
// If checked is true, DNS lookup, bans, and similar are presumed to be already
// checked.
// If forceid is not nil, the function will create the user with that ID if it
// is not in use. if it is, the function will return nil. The caller should be
// prepared for this if it uses forceid, and ensure forceid is a valid UID.
// A new user is not essentially yet "registered"; until they are, they cannot
// communicate or join channels. A user will be considered registered once all
// packages which are holding registration back have permitted it. If checked
// is true, the creator may assume that it is the only package which may be
// holding registration back.
func NewUser(creator string, checked bool, forceid string) (u *User) {

	userMutex.Lock()

	if forceid != "" {
		if users[forceid] == nil {
			u = new(User)
			u.id = forceid
		}
	} else {
		u = new(User)
		for users[getUIDString()] != nil {
			incrementUID()
		}
		u.id = getUIDString()
		incrementUID()
	}
	u.nick = u.id
	users[u.id] = u
	usersByNick[strings.ToUpper(u.nick)] = u

	userMutex.Unlock()

	if u == nil {
		return
	}

	u.checked = checked
	u.regstate = holdRegistration[creator]
	if !checked {
		u.regstate += holdRegistration[""]
	}

	runUserAddHooks(u, creator)
	if u.Registered() {
		runUserRegisterHooks(u)
	}

	return
}

// GetUser gets a user with the given ID, returning a pointer to their User
// structure.
func GetUser(id string) (u *User) {
	userMutex.Lock()
	u = users[id]
	userMutex.Unlock()
	return
}

// GetUserByNick gets a user with the given nick, returning a pointer to their
// User structure.
func GetUserByNick(nick string) (u *User) {
	nick = strings.ToUpper(nick)
	userMutex.Lock()
	u = usersByNick[nick]
	userMutex.Unlock()
	return
}


// Checked returns whether the user is pre-checked for ban purposes, and all
// relevant information added to their data by their creator, and will have no
// holds on registration but setting a nick outside their creating module.
// This can be used to bypass DNS resolution, ban and DNSBL checks, and such.
func (u *User) Checked() bool {
	// As this is just a boolean, we can't possibly receive half-reads, and
	// thus have no need to mutex reads to it.
	return u.checked
}

// ID returns the user's ID.
func (u *User) ID() string {
	// As this is constant, we have no need to mutex reads to it.
	return u.id
}

// SetNick sets the user's nick. This may fail if the nickname is in use.
// If successful, err is nil. If not, err is a message why.
func (u *User) SetNick(nick string) (err os.Error) {
	var oldnick string

	userMutex.Lock()
	u.mutex.Lock()

	oldnick = u.nick
	NICK := strings.ToUpper(nick)

	if nick != oldnick {
		conflict := usersByNick[NICK]
		if conflict != nil && conflict != u {
			err = os.NewError("Nickname is already in use.")
		} else {
			usersByNick[strings.ToUpper(u.nick)] = nil, false
			u.nick = nick
			usersByNick[NICK] = u
		}
	}

	u.mutex.Unlock()
	userMutex.Unlock()

	if err == nil && oldnick != nick {
		runUserNickChangeHooks(u, oldnick, nick)
	}

	return
}

// Nick returns the user's nick.
func (u *User) Nick() (nick string) {
	u.mutex.Lock()
	nick = u.nick
	u.mutex.Unlock()
	return
}

// PermitRegistration marks the user as permitted to register.
// For a user to register, this method must be called a number of times equal
// to that which applicable HoldRegistration() calls were made during init.
func (u *User) PermitRegistration() {
	var registered bool

	u.mutex.Lock()
	if u.regstate > 0 {
		u.regstate--
		if u.regstate == 0 {
			registered = true
		}
	}
	u.mutex.Unlock()

	if registered {
		runUserRegisterHooks(u)
	}
}

// Registered returns whether the user is registered or not.
// Deleted users also count as unregistered.
func (u *User) Registered() (r bool) {
	u.mutex.Lock()
	r = (u.regstate == 0)
	u.mutex.Unlock()
	return
}

// SetData sets the given single piece of metadata on the user.
// source may be nil, in which case the metadata is set by the server.
// Setting it to "" unsets it.
func (u *User) SetData(source *User, name string, value string) {
	var oldvalue string

	u.mutex.Lock()
	if value != "" {
		oldvalue = u.data.Add(name, value)
	} else {
		oldvalue = u.data.Del(name)
	}
	u.mutex.Unlock()

	// If nothing changed, don't call hooks.
	if oldvalue == value {
		return
	}

	c := new(DataChange)
	c.Name = name
	c.Data = value
	old := new(OldData)
	old.Data = value
	runUserDataChangeHooks(source, u, name, oldvalue, value)
	runUserDataChangesHooks(source, u, c, old)
}

// SetDataList performs the given list of metadata changes on the user.
// This is equivalent to lots of SetData calls, except hooks for all data
// changes will receive it as a single list, and it is cheaper.
// source may be nil, in which case the metadata is set by the server.
func (u *User) SetDataList(source *User, c *DataChange) {
	var oldvalues *OldData
	var lasthook *DataChange
	var lastold **OldData = &oldvalues

	u.mutex.Lock()
	for it := c; it != nil; it = it.Next {

		// Make the change.
		var oldvalue string
		if it.Data != "" {
			oldvalue = u.data.Add(it.Name, it.Data)
		} else {
			oldvalue = u.data.Del(it.Name)
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
	u.mutex.Unlock()

	for it, old := c, oldvalues; it != nil && old != nil; it, old = it.Next, old.Next {
		runUserDataChangeHooks(source, u, c.Name, old.Data, it.Data)
	}
	runUserDataChangesHooks(source, u, c, oldvalues)
}

// Data gets the given piece of metadata.
// If it is not set, this method returns "".
func (u *User) Data(name string) (value string) {
	u.mutex.Lock()
	value = u.data.Get(name)
	u.mutex.Unlock()
	return
}

// DataRange calls the given function for every piece of metadata with the
// given prefix. If none are found, the function is never called. Metadata
// items added while this function is running may or may not be missed.
func (u *User) DataRange(prefix string, f func(name, value string)) {
	var dataArray [100]DataChange
	var data []DataChange = dataArray[0:0]
	var root, it trie.StringTrie

	for firstrun := true; firstrun || !it.Empty(); {
		u.mutex.Lock()

		// On our first run, get an iterator at our first value.
		if firstrun {
			root = u.data.GetSub(prefix)
			it = root
			if !it.Empty() {
				if key, _ := it.Value(); key == "" {
					it = it.Next(root)
				}
			}
			firstrun = false
		}

		// Read a block of values.
		for i := 0; !it.Empty() && i < cap(data); i++ {
			data = data[0 : i+1]
			data[i].Name, data[i].Data = it.Value()
			it = it.Next(root)
		}

		u.mutex.Unlock()

		// Call the function for all of them, and clear data.
		for _, item := range data {
			f(item.Name, item.Data)
		}
		data = data[0:0]
	}
}


// Channels returns a pointer to the user's membership list.
func (u *User) Channels() (chans *Membership) {
	u.mutex.Lock()
	chans = u.chans
	u.mutex.Unlock()
	return
}

// Message sends a message directly to the user.
// source may be nil, indicating a message from the server.
// t may be "" (for default), and indicates the type of message.
func (u *User) Message(source *User, message []byte, t string) {

	// Unregistered users may neither send nor receive messages.
	if !u.Registered() || (source != nil && !source.Registered()) {
		return
	}

	// We actually just call hooks, and let the subsystems handle it.
	runUserMessageHooks(source, u, message, t)
}

// Delete deletes the user from the server.
// They are removed from all channels they are in first.
// The given message is recorded as the reason why.
// Source may be nil, to indicate a deletion by the server.
func (u *User) Delete(source *User, message string) {
	var chans *Membership
	var deleted bool

	userMutex.Lock()
	u.mutex.Lock()

	// Delete the user from the user tables.
	if users[u.id] == u {
		users[u.id] = nil, false
	
		NICK := strings.ToUpper(u.nick)
		if usersByNick[NICK] == u {
			usersByNick[NICK] = nil, false
		}
		deleted = true
	}
	userMutex.Unlock()

	// If they were deleted from the user tables, continue.
	if deleted {
 
		// Mark the user as deleted, and get their membership list.
		u.regstate = -1
		chans = u.chans

		// Remove them from all channel lists.
		var prev *Membership
		for it := chans; it != nil; it, prev = it.unext, it {
			u.mutex.Unlock()
			if prev != nil {
				prev.c.mutex.Unlock()
				runChanUserRemoveHooks(it.c.t, u, u, it.c, message)
			}
			it.c.mutex.Lock()
			u.mutex.Lock()

			if it.cprev == nil {
				it.c.users = it.cnext
			} else {
				it.cprev.cnext = it.cnext
			}
			if it.cnext != nil {
				it.cnext.cprev = it.cprev
			}
		}

		u.mutex.Unlock()

		runUserDeleteHooks(source, u, message)
		return
	}

	// Otherwise, unlock them and continue.
	u.mutex.Unlock()
}
