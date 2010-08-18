package core

import "os"
import "strings"


// Represents a user.
type User struct {
	id   string
	nick string
	checked bool
	regcount int
	chans *Membership
	data *Trie
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
		var old interface{}
		if value != "" {
			old = TrieAdd(&u.data, name, value)
		} else {
			old = TrieDel(&u.data, name)
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

			// Make the change.
			var old interface{}
			var oldvalue string
			if it.Data != "" {
				old = TrieAdd(&u.data, it.Name, it.Data)
			} else {
				old = TrieDel(&u.data, it.Name)
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
		runUserDataChangeHooks(source, u, c.Name, old.Data, c.Data)
	}
	runUserDataChangesHooks(source, u, c, oldvalues)
}

// Data gets the given piece of metadata.
// If it is not set, this method returns "".
func (u *User) Data(name string) (value string) {
	wait := make(chan bool)
	corechan <- func() {
		val := TrieGet(&u.data, name)
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
func (u *User) DataRange(prefix string, f func(name, value string)) {
	var dataArray [50]DataChange
	var data []DataChange = dataArray[0:0]
	var root, it *Trie
	wait := make(chan bool)

	// Get an iterator pointing to our first value.
	corechan <- func() {
		root = TrieGetSub(&u.data, prefix)
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


// Channels returns a pointer to the user's membership list.
func (u* User) Channels() (chans *Membership) {
	wait := make(chan bool)
	corechan <- func() {
		chans = u.chans
		wait <- true
	}
	<-wait
	return
}

// Message sends a message directly to the user.
// source may be nil, indicating a message from the server.
// t may be "" (for default), and indicates the type of message.
func (u *User) Message(source *User, message []byte, t string) {

	// Unregistered users may neither send nor receive messages.
	if !u.Registered() || !source.Registered() {
		return
	}

	// We actually just call hooks, and let the subsystems handle it.
	runUserMessageHooks(source, u, message, t)
}

// Delete deletes the user from the server.
// They are removed from all channels they are in first.
// The given message is recorded as the reason why.
func (u *User) Delete(message string) {
	var chans *Membership

	wait := make(chan bool)
	corechan <- func() {
		// Delete the user from the user tables.
		if users[u.id] == u {
			users[u.id] = nil, false
		}
		NICK := strings.ToUpper(u.nick)
		if usersByNick[NICK] == u {
			usersByNick[NICK] = nil, false
		}

		// Remove them from all channel lists.
		chans = u.chans
		for it := u.chans; it != nil; it = it.unext {
			if it.cprev == nil {
				it.c.users = it.cnext
			} else {
				it.cprev.cnext = it.cnext
			}
			if it.cnext != nil {
				it.cnext.cprev = it.cprev
			}
		}

		wait <- true
	}
	<-wait

	for it := chans; it != nil; it = it.UserNext() {
		runChanUserRemoveHooks(it.c.t, u, u, it.c)
	}

	runUserDeleteHooks(u, message)
}

// GetIdent gets an ident for a user, substituting "." if none exists.
func (u *User) GetIdent() (ident string) {
	ident = u.Data("ident")
	if ident == "" {
		ident = "-"
	}
	return
}

// GetHostname gets a hostname for a user, substituting the server name if none
// exists.
func (u *User) GetHostname() (hostname string) {
	hostname = u.Data("hostname")
	if hostname == "" {
		hostname = "Server.name"
	}
	return
}

// GetSetBy gets a string representing who set a piece of metadata set by the
// user. It returns the user's nick!ident@host if they are logged out, and
// their account name if they are logged in.
func (u *User) GetSetBy() (setby string) {
	wait := make(chan bool)
	corechan <- func() {
		if v := TrieGet(&u.data, "account"); v != nil {
			setby = v.(string)
		} else {
			ident := "-"
			if v := TrieGet(&u.data, "ident"); v != nil {
				ident = v.(string)
			}
			hostname := "Server.name"
			if v := TrieGet(&u.data, "hostname"); v != nil {
				hostname = v.(string)
			}
			setby = u.nick + "!" + ident + "@" + hostname
		}
		wait <- true
	}
	<-wait
	return
}

// GetDecentBan gets a reasonable default ban string on this user, without
// a ban, ban exception, or unrestriction prefix attached.
// At present, it uses an account ban if they are logged into an account, or
// a host ban on their ident@host otherwise. As such, it should be used only
// if a means to adminstrate both kinds of ban is available.
func (u *User) GetDecentBan() string {

	// Logged in? Ban them by account.
	if v := u.Data("account"); v != "" {
		return "account " + v
	}

	ident := u.GetIdent()
	hostname := u.GetHostname()

	if ident[0] == '~' {
		ident = ident[1:]
	}

	if v := strings.IndexRune(hostname, '.'); v != -1 && len(hostname) > v+1 {
		// Decent for all resolved hosts.
		if hostname != u.Data("ip") {
			hostname = "*" + hostname[v+1:]
		} else {
			// Decent for IPv4.
			v = strings.LastIndex(hostname, ".")
			hostname = hostname[0:v+1] + "*"
		}
	} else if v := strings.IndexRune(hostname, ':'); v != -1 && len(hostname) > v+1 {
		// Decent for IPv6. Maybe.
		// It'll do until a better idea shows up.
		v = strings.LastIndex(hostname, ":")
		hostname = hostname[0:v+1] + "*"
	}

	return "host *!~" + ident + "@" + hostname
}
