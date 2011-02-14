package core

import "os"
import "strings"
import "sync"
import "time"
import "unsafe"


import "oddcomm/lib/trie"


// Represents a user.
type User struct {
	mutex    sync.Mutex
	id       string
	nick     string
	nickts   int64
	checked  bool
	regstate int
	chans    *Membership
	data     trie.StringTrie
	owner    string
	owndata  interface{}
}

// IterateUsers calls the given function for all existing users.
// If owner is not "", it is only called for users with that owner.
//
// Users added concurrently may or may not be missed.
func IterateUsers(owner string, f func(u *User)) {
	it := users.Iterate()
	if it == nil {
		return
	}

	if owner != "" {
		_, uptr := it.Value()
		u := (*User)(uptr)
		if u.owner == owner {
			f(u)
		}
		for it.Next() {
			_, uptr = it.Value()
			u = (*User)(uptr)
			if u.owner == owner {
				f(u)
			}
		}
	} else {
		_, uptr := it.Value()
		f((*User)(uptr))
		for it.Next() {
			_, uptr = it.Value()
			f((*User)(uptr))
		}
	}
}


// NewUser creates a new user.
//
// owner is the name of the package owning the user, and owndata is arbitrary
// data associated with the user by that package for its own usage.
//
// If checked is true, DNS lookup, bans, and similar are presumed to be already
// checked for this user. This generally is used if the user is added remotely.
//
// If forceid is not nil, the function will create the user with that ID if it
// is not in use. if it is, the function will return nil. The caller should be
// prepared for this if it uses forceid, and ensure forceid is a valid UID.
//
// A new user is not essentially yet "registered"; until they are, they cannot
// communicate or join channels. A user will be considered registered once all
// packages which are holding registration back have permitted it. If checked
// is true, the owner may assume that it is the only package which may be
// holding registration back.
//
// data is a list of DataChanges to apply to the new user immediately. This
// should generally include IP and ident if the user has them.
//
// They will be applied prior to the new user hooks being called, with
// usual data change hooks called after. It may be nil.
func NewUser(owner string, owndata interface{}, checked bool, forceid string, data []DataChange) (u *User) {
	done := make([]DataChange, 0, len(data))
	old := make([]string, 0, len(data))

	u = new(User)

	userMutex.Lock()

	// Figure out what ID the user should have.
	if forceid != "" {
		if users.Get(forceid) == nil {
			u.id = forceid
		} else {
			u = nil
		}
	} else {
		for users.Get(getUIDString()) != nil {
			incrementUID()
		}
		u.id = getUIDString()
		incrementUID()
	}

	// If they have a valid ID, set them up and add them.
	// As reads occur without holding a mutex, they should be fully set up
	// before being added so we're never not in a safe state.
	if u != nil {
		u.mutex.Lock()

		// Set their nick, owner, and owndata.
		u.nick = u.id
		u.owner = owner
		u.owndata = owndata

		// Set the number of permits this user should wait for before
		// completing registration.
		u.checked = checked
		u.regstate = 1 + holdRegistration[owner]
		if !checked {
			u.regstate += holdRegistration[""]
		}

		// Apply provided data to the user.
		for _, it := range data {

			// Make the change.
			var oldvalue string
			if it.Data != "" {
				oldvalue = u.data.Insert(it.Name, it.Data)
			} else {
				oldvalue = u.data.Remove(it.Name)
			}

			// If this was a do-nothing change, do not report it.
			if oldvalue == it.Data {
				continue
			}

			// Add to slices to send to hooks.
			// Still need to track old data, just in case someone
			// decides to set a value TWICE for some reason.
			done = append(done, it)
			old = append(old, oldvalue)
		}

		// Now fully configured, add them.
		users.Insert(u.id, unsafe.Pointer(u))
		usersByNick.Insert(u.nick, unsafe.Pointer(u))
	} else {
		userMutex.Unlock()
		return
	}

	// Run user addition and data change hooks.
	hookRunner <- func() {
		runUserAddHooks(owner, u)
		for i, it := range done {
			runUserDataChangeHooks(owner, nil, u, it.Name, old[i], it.Data)
		}
	}

	u.mutex.Unlock()
	userMutex.Unlock()
	return
}

// GetUser gets a user with the given ID, returning a pointer to their User
// structure.
func GetUser(id string) (u *User) {
	return (*User)(users.Get(id))
}

// GetUserByNick gets a user with the given nick, returning a pointer to their
// User structure.
func GetUserByNick(nick string) (u *User) {
	NICK := strings.ToUpper(nick)
	u = (*User)(usersByNick.Get(NICK))
	return
}

// Checked returns whether the user is pre-checked for ban purposes, and all
// relevant information added to their data by their creator, and will have no
// holds on registration outside their creating module.
// This can be used to bypass DNS resolution, ban and DNSBL checks, and such.
func (u *User) Checked() bool {
	return u.checked
}

// ID returns the user's ID.
func (u *User) ID() string {
	return u.id
}

// SetNick sets the user's nick. This may fail if the nickname is in use.
// If ts is not -1, it sets the nick timestamp to store.
// If successful, err is nil. If not, err is a message why.
func (u *User) SetNick(pkg, nick string, ts int64) (err os.Error) {
	if ts == -1 {
		ts = time.Seconds()
	}

	u.mutex.Lock()

	oldnick := u.nick
	OLDNICK := strings.ToUpper(oldnick)
	NICK := strings.ToUpper(nick)

	if OLDNICK != NICK {
		userMutex.Lock()

		// Change the nick if there's no conflict.
		conflict := (*User)(usersByNick.Get(NICK))
		if conflict != nil && conflict != u {
			err = os.NewError("Nickname is already in use.")
		} else {
			usersByNick.Remove(strings.ToUpper(oldnick))
			usersByNick.Insert(NICK, unsafe.Pointer(u))
			u.nick = nick
			u.nickts = ts
		}

		userMutex.Unlock()

	} else if oldnick != nick {
		// Just a change in case; no need to alter the tables,
		// or change nick timestamp.
		u.nick = nick
		ts = u.nickts
	}

	if oldnick != nick {
		if err == nil {
			hookRunner <- func() {
				runUserNickChangeHooks(pkg, u, oldnick, nick, ts)
			}
		}
	}

	u.mutex.Unlock()

	return
}

// Nick returns the user's nick.
func (u *User) Nick() (nick string) {
	return u.nick
}

// NickTS returns the user's nick timestamp.
func (u *User) NickTS() (ts int64) {
	return u.nickts
}

// PermitRegistration marks the user as permitted to register.
// For a user to register, this method must be called a number of times equal
// to that which applicable HoldRegistration() calls were made during init,
// PLUS once from the code creating the user whenever it has finished setting
// it up.
func (u *User) PermitRegistration(pkg string) {
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
			runUserRegisterHooks(pkg, u)
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
func (u *User) SetData(pkg string, source *User, name, value string) {

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

	var c DataChange
	c.Name = name
	c.Data = value
	hookRunner <- func() {
		runUserDataChangeHooks(pkg, source, u, name, oldvalue, value)
		runUserDataChangesHooks(pkg, source, u, []DataChange{c}, []string{oldvalue})
	}

	u.mutex.Unlock()
}

// SetDataList performs the given list of metadata changes on the user.
// This is equivalent to lots of SetData calls, except hooks for all data
// changes will receive it as a single list, and it is cheaper.
// source may be nil, in which case the metadata is set by the server.
func (u *User) SetDataList(pkg string, source *User, changes []DataChange) {
	done := make([]DataChange, 0, len(changes))
	old := make([]string, 0, len(changes))

	// We hold this so hooks are sent to the hook runner in the order that
	// the data changes were made.
	u.mutex.Lock()

	for _, it := range changes {

		// Make the change.
		var oldvalue string
		if it.Data != "" {
			oldvalue = u.data.Insert(it.Name, it.Data)
		} else {
			oldvalue = u.data.Remove(it.Name)
		}

		// If this was a do-nothing change, don't report it.
		if oldvalue == it.Data {
			continue
		}

		// Add to slices to pass to hooks.
		done = append(done, it)
		old = append(old, oldvalue)
	}

	hookRunner <- func() {
		for i, it := range done {
			runUserDataChangeHooks(pkg, source, u, it.Name, old[i], it.Data)
		}
		runUserDataChangesHooks(pkg, source, u, done, old)
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


// Owner returns the owning package of the user, if one is set.
func (u *User) Owner() string {
	return u.owner
}

// Owndata returns the owning package of the user's data.
func (u *User) Owndata() interface{} {
	return u.owndata
}


// Message sends a message directly to the user.
// source may be nil, indicating a message from the server.
// t may be "" (for default), and indicates the type of message.
func (u *User) Message(pkg string, source *User, message []byte, t string) {

	// Unregistered users may neither send nor receive messages.
	if !u.Registered() || (source != nil && !source.Registered()) {
		return
	}

	// We actually just call hooks, and let the subsystems handle it.
	runUserMessageHooks(pkg, source, u, message, t)
}

// Delete deletes the user from the server.
// They are removed from all channels they are in first.
// The given message is recorded as the reason why.
// Source may be nil, to indicate a deletion by the server.
func (u *User) Delete(pkg string, source *User, message string) {
	var chans *Membership

	userMutex.Lock()
	u.mutex.Lock()

	// This check ensures we're not called more than once.
	if u.regstate == -1 {
		u.mutex.Unlock()
		userMutex.Unlock()
		return
	}

	// Delete the user from the user tables.
	NICK := strings.ToUpper(u.nick)
	users.Remove(u.id)
	usersByNick.Remove(NICK)


	// Mark the user as deleted, and get their membership list.
	regged := (u.regstate == 0)
	u.regstate = -1

	// Run on delete hooks, then on remove hooks for all channels the user
	// is in, at the same time removing them from the channel lists.
	// This ensures the user is still in their channels when deletion hooks
	// are run.
	hookRunner <- func() {
		runUserDeleteHooks(pkg, source, u, message, regged)

		u.mutex.Lock()

		// Get the user's channel list, wiping it here in the process.
		chans = u.chans
		u.chans = nil

		// Remove them from all channel lists and run hooks.
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
					runChanUserRemoveHooks(pkg, prev.c.t, u, u, prev.c, message)
				}
			}
			it.c.mutex.Unlock()
		}

		u.mutex.Unlock()
	}

	u.mutex.Unlock()
	userMutex.Unlock()
}
