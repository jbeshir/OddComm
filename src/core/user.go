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
//
// If forceid is not nil, the function will create the user with that ID if it
// is not in use. if it is, the function will return nil. The caller should be
// prepared for this if it uses forceid, and ensure forceid is a valid UID.
//
// A new user is not essentially yet "registered"; until they are, they cannot
// communicate or join channels. A user will be considered registered once all
// packages which are holding registration back have permitted it. If checked
// is true, the creator may assume that it is the only package which may be
// holding registration back.
//
// data is a list of DataChanges to apply to the new user immediately.
// They will be applied prior to the new user hooks being called, with
// usual data change hooks called after. It may be nil.
func NewUser(creator string, checked bool, forceid string, data *DataChange) (u *User) {
	u = new(User)

	userMutex.Lock()

	// Figure out what ID the user should have.
	if forceid != "" {
		if users[forceid] == nil {
			u.id = forceid
		} else {
			u = nil
		}
	} else {
		for users[getUIDString()] != nil {
			incrementUID()
		}
		u.id = getUIDString()
		incrementUID()
	}

	// If they have a valid ID, set their nick and add them to the user
	// structure.
	if u != nil {
		u.nick = u.id
		users[u.id] = u
		usersByNick[strings.ToUpper(u.nick)] = u
	}

	if u == nil {
		userMutex.Unlock()
		return
	}

	u.mutex.Lock()

	// Set the number of permits this user should wait for before
	// completing registration.
	u.checked = checked
	u.regstate = 1 + holdRegistration[creator]
	if !checked {
		u.regstate += holdRegistration[""]
	}

	// Apply provided data to the user.
	var oldvalues *OldData
	var lasthook *DataChange
	var lastold **OldData = &oldvalues
	for it := data; it != nil; it = it.Next {

		// Make the change.
		var oldvalue string
		if it.Data != "" {
			oldvalue = u.data.Insert(it.Name, it.Data)
		} else {
			oldvalue = u.data.Remove(it.Name)
		}

		// If this was a do-nothing change, cut it out.
		if oldvalue == it.Data {
			if lasthook != nil {
				lasthook.Next = it.Next
			} else {
				data = it.Next
			}
			continue
		}

		// Still need to track old data, just in case someone decides
		// to set a value TWICE for some reason.
		olddata := new(OldData)
		olddata.Data = oldvalue
		*lastold = olddata
		lasthook = it
		lastold = &olddata.Next
	}


	// Run user addition and data change hooks.
	hookRunner <- func() {
		runUserAddHooks(u, creator)
		for it, old := data, oldvalues; it != nil && old != nil;
				it, old = it.Next, old.Next {
			runUserDataChangeHooks(nil, u, it.Name, old.Data, it.Data)
		}
	}

	u.mutex.Unlock()
	userMutex.Unlock()
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

	if err == nil && oldnick != nick {
		hookRunner <- func() {
			runUserNickChangeHooks(u, oldnick, nick)
		}
	}

	u.mutex.Unlock()
	userMutex.Unlock()

	return
}

// Nick returns the user's nick.
func (u *User) Nick() (nick string) {
	return u.nick
}

// PermitRegistration marks the user as permitted to register.
// For a user to register, this method must be called a number of times equal
// to that which applicable HoldRegistration() calls were made during init,
// PLUS once from the code creating the user whenever it has finished setting
// it up.
func (u *User) PermitRegistration() {
	var registered bool

	u.mutex.Lock()
	if u.regstate > 0 {
		u.regstate--
		if u.regstate == 0 {
			registered = true
		}
	}

	if registered {
		hookRunner <- func() {
			runUserRegisterHooks(u)
		}
	}

	u.mutex.Unlock()
}

// Registered returns whether the user is registered or not.
// Deleted users also count as unregistered.
func (u *User) Registered() (r bool) {
	return (u.regstate == 0)
}

// SetData sets the given single piece of metadata on the user.
// source may be nil, in which case the metadata is set by the server.
// Setting it to "" unsets it.
func (u *User) SetData(source *User, name string, value string) {

	// We hold this so hooks are sent to the hook runner in the order that
	// the data changes were made.
	u.mutex.Lock()

	var oldvalue string

	if value != "" {
		oldvalue = u.data.Insert(name, value)
	} else {
		oldvalue = u.data.Remove(name)
	}

	// If nothing changed, don't call hooks.
	if oldvalue == value {
		return
	}

	c := new(DataChange)
	c.Name = name
	c.Data = value
	old := new(OldData)
	old.Data = value
	hookRunner <- func() {
		runUserDataChangeHooks(source, u, name, oldvalue, value)
		runUserDataChangesHooks(source, u, c, old)
	}

	u.mutex.Unlock()
}

// SetDataList performs the given list of metadata changes on the user.
// This is equivalent to lots of SetData calls, except hooks for all data
// changes will receive it as a single list, and it is cheaper.
// source may be nil, in which case the metadata is set by the server.
func (u *User) SetDataList(source *User, c *DataChange) {
	var oldvalues *OldData
	var lasthook *DataChange
	var lastold **OldData = &oldvalues

	// We hold this so hooks are sent to the hook runner in the order that
	// the data changes were made.
	u.mutex.Lock()

	for it := c; it != nil; it = it.Next {

		// Make the change.
		var oldvalue string
		if it.Data != "" {
			oldvalue = u.data.Insert(it.Name, it.Data)
		} else {
			oldvalue = u.data.Remove(it.Name)
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

	hookRunner <- func() {
		for it, old := c, oldvalues; it != nil && old != nil;
				it, old = it.Next, old.Next {
			runUserDataChangeHooks(source, u, it.Name, old.Data, it.Data)
		}
		runUserDataChangesHooks(source, u, c, oldvalues)
	}

	u.mutex.Unlock()
}

// Data gets the given piece of metadata.
// If it is not set, this method returns "".
func (u *User) Data(name string) (value string) {
	value = u.data.Get(name)
	return
}

// DataRange calls the given function for every piece of metadata with the
// given prefix. If none are found, the function is never called. Metadata
// items added while this function is running may or may not be missed.
func (u *User) DataRange(prefix string, f func(name, value string)) {
	for it := u.data.GetSub(prefix); it != nil; {
		name, data := it.Value()
		f(name, data)
		if !it.Next() {
			it = nil
		}
	}
}


// Channels returns a pointer to the user's membership list.
func (u *User) Channels() (chans *Membership) {
	chans = u.chans
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

	// Holding the global user mutex for hook ordering throughout
	// also means we know we're not called concurrently.
	userMutex.Lock()
	u.mutex.Lock()

	// And this check ensures we're not called more than once.
	if u.regstate == -1 {
		u.mutex.Unlock()
		userMutex.Unlock()
		return
	}

	// Delete the user from the user tables.
	NICK := strings.ToUpper(u.nick)
	users[u.id] = nil, false
	usersByNick[NICK] = nil, false

	// Mark the user as deleted, and get their membership list.
	u.regstate = -1
	chans = u.chans
	u.chans = nil

	// This is required by the ordering on mutexes, but means we
	// must be able to otherwise assume that deleted users will not
	// have their channel membership written to.
	// We ensure this elsewhere.
	u.mutex.Unlock()

	// Remove them from all channel lists.
	var prev, it *Membership
	for it = chans; it != nil; it = it.unext {

		u.mutex.Unlock()
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

		prev = it
		if prev != nil {
			hookRunner <- func() {
				runChanUserRemoveHooks(prev.c.t, u, u, prev.c, message)
			}
		}
		it.c.mutex.Unlock()
	}

	hookRunner <- func() {
		runUserDeleteHooks(source, u, message)
	}

	userMutex.Unlock()
}
