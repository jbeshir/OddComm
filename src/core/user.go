package core

import "os"
import "strings"


// Represents a user.
type User struct {
	id   string
	nick string
	checked bool
	regcount int
	data map[string]string
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
	wait := make(chan bool)
	corechan <- func() {
		if forceid != "" && users[forceid] != nil {
			wait <- true
			return
		}

		u = new(User)
		u.data = make(map[string]string)
		u.checked = checked
		u.regcount = holdRegistration[creator]
		if (!checked) {
			u.regcount += holdRegistration[""]
		}

		if forceid != "" {
			u.id = forceid
		} else {
			for users[getUIDString()] != nil {
				incrementUID()
			}
			u.id = getUIDString()
			incrementUID()
		}

		u.nick = u.id
		users[u.id] = u
		usersByNick[strings.ToUpper(u.nick)] = u
		wait <- true
	}
	<-wait

	runUserAddHooks(u, creator)
	if (u.Registered()) {
		runUserRegisterHooks(u)
	}

	return
}

// GetUser gets a user with the given ID, returning a pointer to their User
// structure.
func GetUser(id string) *User {
	c := make(chan *User)
	corechan <- func() {
		c <- users[id]
	}
	return <-c
}

// GetUserByNick gets a user with the given nick, returning a pointer to their
// User structure.
func GetUserByNick(nick string) *User {
	c := make(chan *User)
	corechan <- func() {
		c <- usersByNick[strings.ToUpper(nick)]
	}
	return <-c
}


// Checked returns whether the user is pre-checked for ban purposes, and all
// relevant information added to their data by their creator, and will have no
// holds on registration but setting a nick outside their creating module.
// This can be used to bypass DNS resolution, ban and DNSBL checks, and such.
func (u *User) Checked() bool {
	return u.checked
}

// ID returns the user's ID.
func (u *User) ID() (id string) {
	wait := make(chan bool)
	corechan <- func() {
		id = u.id
		wait <- true
	}
	<-wait

	return
}

// SetNick sets the user's nick. This may fail if the nickname is in use.
// If successful, err is nil. If not, err is a message why.
func (u *User) SetNick(nick string) (err os.Error) {
	var oldnick string

	wait := make(chan bool)
	corechan <- func() {
		oldnick = u.nick
		NICK := strings.ToUpper(nick)

		if nick == oldnick {
			wait <- true
			return
		}

		conflict := usersByNick[NICK]
		if conflict != nil && conflict != u {
			err = os.NewError("Nickname is already in use.")
			wait <- true
			return
		}

		usersByNick[strings.ToUpper(u.nick)] = nil, false
		u.nick = nick
		usersByNick[NICK] = u
		
		wait <- true
	}
	<-wait

	if err == nil && oldnick != nick {
		runUserNickChangeHooks(u, oldnick, nick)
	}

	return
}

// Nick returns the user's nick.
func (u *User) Nick() (nick string) {
	wait := make(chan bool)
	corechan <- func() {
		nick = u.nick
		wait <- true
	}
	<-wait

	return
}

// PermitRegistration marks the user as permitted to register.
// For a user to register, this method must be called a number of times equal
// to that which applicable HoldRegistration() calls were made during init.
func (u *User) PermitRegistration() {
	c := make(chan bool)
	corechan <- func() {
		if u.regcount <= 0 {
			c <- false
			return
		}

		u.regcount--

		if u.regcount == 0 {
			c <- true
		} else {
			c <- false
		}
	}
	registered := <-c

	if registered {
		runUserRegisterHooks(u)
	}
}

// Registered returns whether the user is registered or not.
func (u *User) Registered() bool {
	c := make(chan bool)
	corechan <- func() {
		c <- (u.regcount == 0)
	}
	return <-c
}

// SetData sets the given single piece of metadata on the user.
// source may be nil, in which case the metadata is set by the server.
// Setting it to "" unsets it.
func (u *User) SetData(source *User, name string, value string) {
	var oldvalue string

	wait := make(chan bool)
	corechan <- func() {
		oldvalue = u.data[name]
		if value != "" {
			u.data[name] = value
		} else {
			u.data[name] = "", false
		}

		wait <- true
	}
	<-wait

	// If nothing changed, don't call hooks.
	if oldvalue == value {
		return
	}

	runUserDataChangeHooks(source, u,name, oldvalue, value)

	c := new(DataChange)
	c.Name = name
	c.Data = value
	old := new(OldData)
	old.Data = value
	runUserDataChangesHooks(source, u, c, old)
}

// SetDataList performs the given list of metadata changes on the user.
// This is equivalent to lots of SetData calls, except hooks for all data
// changes will receive it as a single list, and it is cheaper.
// source may be nil, in which case the metadata is set by the server.
func (u *User) SetDataList(source *User, c *DataChange) {
	var oldvalues *OldData
	wait := make(chan bool)
	corechan <- func() {
		var lasthook *DataChange
		for it := c; it != nil; it = it.Next {

			// If this is a do-nothing change, cut it out.
			if u.data[it.Name] == it.Data {
				if lasthook != nil {
					lasthook.Next = it.Next
				} else {
					c = it.Next
				}
			}

			old := new(OldData)
			old.Data = u.data[it.Name]
			old.Next = oldvalues
			oldvalues = old

			if it.Data != "" {
				u.data[it.Name] = it.Data
			} else {
				u.data[it.Name] = "", false
			}

			lasthook = it
		}

		wait <- true
	}
	<-wait

	for it, old := c, oldvalues; it != nil && old != nil; it, old = it.Next, old.Next {
		runUserDataChangeHooks(source, u, c.Name, old.Data, c.Data)
	}
	runUserDataChangesHooks(source, u, c, oldvalues)
}

// Data gets the given piece of metadata.
// If it is not set, this method returns "".
func (u *User) Data(name string) (value string) {
	wait := make(chan bool)
	corechan <- func() {
		value = u.data[name]
		wait <- true
	}
	<-wait

	return
}

// PM sends a message directly to the user.
// source may be nil, indicating a message from the server.
// t may be "" (for default), and indicates the type of message.
func (u *User) PM(source *User, message []byte, t string) {

	// Unregistered users may neither send nor receive messages.
	if !u.Registered() || !source.Registered() {
		return
	}

	// We actually just call hooks, and let the subsystems handle it.
	runUserPMHooks(source, u, message, t)
}

// Delete deletes the user from the server.
// The given message is recorded as the reason why.
func (u *User) Delete(message string) {
	wait := make(chan bool)
	corechan <- func() {
		if users[u.id] == u {
			users[u.id] = nil, false
		}
		if users[u.nick] == u {
			usersByNick[u.nick] = nil, false
		}

		wait <- true
	}
	<-wait

	runUserDeleteHooks(u, message)
}
