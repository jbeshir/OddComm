package core

import "os"


type User struct {
	id   string
	nick string
	regcount int
	data map[string]string
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

// SetNick sets the user's nick.
// If successful, err is nil. If not, err is a message why.
func (u *User) SetNick(nick string) (err os.Error) {
	var oldnick string

	wait := make(chan bool)
	corechan <- func() {
		oldnick = u.nick

		if usersByNick[nick] == u {
			wait <- true
			return
		} else 	if usersByNick[nick] != nil {
			err = os.NewError("already in use")
			wait <- true
			return
		}

		usersByNick[u.nick] = nil, false
		u.nick = nick
		usersByNick[u.nick] = u
		
		wait <- true
	}
	<-wait

	if err == nil && oldnick != nick {
		runNickChangeHooks(u, oldnick)
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

// Decrements the number of modules who still need to sign off before this
// user is registered.
func (u *User) DecrementRegcount() {
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
		runRegisteredHooks(u)
	}
}

// Registered returns whether the user is registered yet or not.
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

	runDataChangeHooks(u, name, oldvalue)
}

// Data gets the given piece of metadata.
// If it is not set, it will be "".
func (u *User) Data(name string) (value string) {
	wait := make(chan bool)
	corechan <- func() {
		value = u.data[name]
		wait <- true
	}
	<-wait

	return
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

	runRemovedHooks(u, message)
}
