package core

import "os"
import "strings"


type User struct {
	id   string
	nick string
	checked bool
	regcount int
	data map[string]string
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

// SetData sets the given piece of metadata on the user.
// Setting it to "" unsets it.
func (u *User) SetData(name string, value string) {
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

	runUserDataChangeHooks(u, name, oldvalue, value)
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
func (u *User) PM(source *User, message, t string) {

	// Unregistered users may neither send nor receive messages.
	if !u.Registered() || !source.Registered() {
		return
	}

	// We actually just call hooks, and let the subsystems handle it.
	runUserPMHooks(source, u, message, t)
}

// Remove kills the user.
// The given message is recorded as the reason why.
func (u *User) Remove(message string) {
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

	runUserRemovedHooks(u, message)
}
