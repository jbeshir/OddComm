/*
	The core of the server.

	Tracks users and channels, and their associations. Exports methods to
	do things with them.
*/
package core

import "os"


// The main users by ID map to look up users.
var users map[string]*CoreUser

// The users by nick map for indexed look up of users by name.
var usersByNick map[string]*CoreUser

// The package message channel by package name map.
// For sending messages to package goroutines.
var packages map[string]chan string

// The core channel.
// For sending functions to to run in the goroutine that owns the core's
// data structures.
var corechan chan func()


func init() {
}


// Core initialisation function.
// Initialises variables and starts the core goroutine, which does not need
// to do cleanup before shutdown because it is always ready to stop whenever
// everything else is done calling into it.
func init() {
	users = make(map[string]*CoreUser)
	usersByNick = make(map[string]*CoreUser)
	packages = make(map[string]chan string)

	corechan = make(chan func())
	go func() {
		for corefunc := range corechan {
			corefunc()
		}
	}()
}


// AddPackage adds the given package to the package list.
// Packages should use a unique name.
func AddPackage(name string, c chan string) {
	wait := make(chan bool)
	corechan <- func() {
		packages[name] = c
		wait <- true
	}
	<-wait
}


// Shutdown shuts down the server by sending an exit message to every package
// goroutine. This is done asynchronously.
func Shutdown() {
	go func() {
		for name := range packages {
			packages[name] <- "exit"
		}
	}()
}


// AddUser adds a user with the given name, returning a pointer to its User
// structure.
func AddUser(nick string) (u *CoreUser, err os.Error) {
	wait := make(chan bool)
	corechan <- func() {
		if usersByNick[nick] != nil {
			err = os.NewError("already in use")
			wait <- true
			return
		}

		u = new(CoreUser)
		u.id = "1"
		u.nick = nick
		users[u.id] = u
		usersByNick[u.nick] = u
		wait <- true
	}
	<-wait

	return
}

// GetUser gets a user with the given ID, returning a pointer to their User
// structure.
func GetUser(id string) *CoreUser {
	c := make(chan *CoreUser)
	corechan <- func() {
		c <- users[id]
	}
	return <-c
}

// GetUserByNick gets a user with the given nick, returning a pointer to their
// User structure.
func GetUserByNick(nick string) *CoreUser {
	c := make(chan *CoreUser)
	corechan <- func() {
		c <- users[nick]
	}
	return <-c
}
