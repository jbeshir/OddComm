/*
	The core of the server.

	Tracks users and channels, and their associations. Exports methods to
	do things with them.
*/
package core


// The main users by ID map to look up users.
var users map[string]*User

// The users by nick map for indexed look up of users by name.
var usersByNick map[string]*User

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
	users = make(map[string]*User)
	usersByNick = make(map[string]*User)
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


// NewUser creates a new user.
func NewUser() (u *User) {
	wait := make(chan bool)
	corechan <- func() {
		u = new(User)
		u.data = make(map[string]string)
		u.id = "1"
		u.regcount = initialRegcount
		users[u.id] = u
		wait <- true
	}
	<-wait

	runNewUserHooks(u)

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
		c <- users[nick]
	}
	return <-c
}
