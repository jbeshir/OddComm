/*
	The core of the server.

	Tracks users and channels, and their associations. Exports methods to
	do things with them.
*/
package core

// Sets the server version strinhg.
var Version string = "0.0.1"


// The main users by ID map to look up users.
var users map[string]*User

// The users by nick map for indexed look up of users by name.
var usersByNick map[string]*User

// The main channels by type, by name map to look up channels.
var channels map[string]map[string]*Channel

// The package message channel by package name map.
// For sending messages to package goroutines.
var packages map[string]chan string

// The core channel.
// For sending functions to to run in the goroutine that owns the core's
// data structures.
var corechan chan func()


// Core initialisation function.
// Initialises variables and starts the core goroutine, which does not need
// to do cleanup before shutdown because it is always ready to stop whenever
// everything else is done calling into it.
func init() {
	users = make(map[string]*User)
	usersByNick = make(map[string]*User)
	channels = make(map[string]map[string]*Channel)
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
