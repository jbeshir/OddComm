package core

import "os"


type User struct {
	id   string
	nick string
	regcount int
	data map[string]string
}


// ID returns the user's ID.
func (u *User) ID() string {
	c := make(chan string)
	corechan <- func() {
		c <- u.id
	}
	return <-c
}

// SetNick sets the user's nick.
// If successful, err is nil. If not, err is a message why.
func (u *User) SetNick(nick string) (err os.Error) {
	wait := make(chan bool)
	corechan <- func() {
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
	}
	<-wait

	return
}

// Nick returns the user's nick.
func (u *User) Nick() string {
	c := make(chan string)
	corechan <- func() {
		c <- u.nick
	}
	return <-c
}

// SetData sets the given piece of metadata on the user.
// Setting it to "" unsets it.
func (u *User) SetData(name string, value string) {
	wait := make(chan bool)
	corechan <- func() {
		if value != "" {
			u.data[name] = value
		} else {
			u.data[name] = "", false
		}
	}
	<-wait
}

// Data gets the given piece of metadata.
// If it is not set, it will be "".
func (u *User) Data(name string) (value string) {
	wait := make(chan bool)
	corechan <- func() {
		value = u.data[name]
	}
	<-wait

	return
}

// Remove kills the user.
// The given message is recorded as the reason why.
func (u *User) Remove(_ string) {
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
}
