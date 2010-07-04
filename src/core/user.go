package core

import "os"

type User interface {
	// ID returns the user's ID.
	ID() (id string)
	// Nick returns the user's nick.
	Nick() (nick string)

	// SetNick sets the user's nick.
	// If successful, err is nil. If not, err is a message why.
	SetNick(nick string) (err os.Error)

	// SetData sets the given piece of metadata on the user.
	// Setting it to "" unsets it.
	SetData(name string, value string)

	// GetData gets the given piece of metadata.
	// If it is not set, it will be "".
	GetData(name string) (value string)

	// Remove kills the user.
	// The given message is recorded as the reason why.
	Remove(message []byte)
}

type CoreUser struct {
	id   string
	nick string
	data map[string]string
}


func (u *CoreUser) Nick() string {
	c := make(chan string)
	corechan <- func() {
		c <- u.nick
	}
	return <-c
}

func (u *CoreUser) ID() string {
	c := make(chan string)
	corechan <- func() {
		c <- u.id
	}
	return <-c
}

func (u *CoreUser) SetNick(nick string) (err os.Error) {
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

func (u *CoreUser) SetData(name string, value string) {
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

func (u *CoreUser) GetData(name string) (value string) {
	wait := make(chan bool)
	corechan <- func() {
		value = u.data[name]
	}
	<-wait

	return
}

func (u *CoreUser) Remove(_ []byte) {
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
