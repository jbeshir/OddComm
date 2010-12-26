/*
	The core of the server.

	Tracks users and channels, and their associations. Exports methods to
	do things with them.
*/
package core

import "sync"

// An ordering is imposed on mutexes within this package: Only one global, one
// channel, and one user mutex may be held at once, and they must be acquired
// in that order.
//
// Membership entry changes should only be performed while holding the correct
// channel mutex and the correct user mutex.
//
// Generally, events with hooks that alter state should have a mutex held
// between the alteration and hooks being called, to ensure hooks are called in
// the same order as these events.
//
// While a mutex exists that can block writes to any one global or Extensible
// structure while it is synchronised, no mutex exists to lock all extensible
// structures at once; they would all need to be locked. If this is
// The global mutexes, aside providing the above, also permit changes to these
// data structures to be blocked while they're being synchronised. Individual
// Extensible types have their own mutexes which can be used similarly, 
// The global mutexes serve this function and one other- to permit changes to
// be blocked while serialising the state of the whole server.


// Sets the server version string.
var Version string = "0.0.1"


// The users by ID and users by nick maps to look up users.
var users map[string]*User
var usersByNick map[string]*User
var userMutex sync.Mutex

// The channels by type, by name map to look up channels.
var channels map[string]map[string]*Channel
var chanMutex sync.Mutex

// The package message channel by package name map.
// For sending messages to package goroutines.
var packages map[string]chan string
var packageMutex sync.Mutex


// Core initialisation function.
// Initialises variables and starts the core goroutine, which does not need
// to do cleanup before shutdown because it is always ready to stop whenever
// everything else is done calling into it.
func init() {
	users = make(map[string]*User)
	usersByNick = make(map[string]*User)
	channels = make(map[string]map[string]*Channel)
	packages = make(map[string]chan string)
}


// AddPackage adds the given package to the package list.
// Packages should use a unique name.
func AddPackage(name string, c chan string) {
	packageMutex.Lock()
	packages[name] = c
	packageMutex.Unlock()
}


// Shutdown shuts down the server by sending an exit message to every package
// goroutine. This is done asynchronously.
func Shutdown() {
	go func() {
		packageMutex.Lock()
		for name := range packages {
			packages[name] <- "exit"
		}
		packageMutex.Unlock()
	}()
}
