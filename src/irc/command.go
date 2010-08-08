package irc

import "io"
import "strings"

import "oddircd/src/core"


// Command structure; contains the handler and information for a command.
type Command struct {

	// The command handler.
	// w may be nil.  If non-nil, it writes to the client's connection.
	// This handler may assume the contraints set in the rest of the
	// structure have been met when it is called.
	Handler func(u *core.User, w io.Writer, params [][]byte)

	// The minimum arguments the command expects.
	Minargs int

	// The maximum arguments the command expects.
	// Excess arguments are merged into the last.
	// This has a fixed maximum value of 50.
	Maxargs int

	// Whether this command can be called for unregistered users.
	// 0: Registered users only.
	// 1: Both registered users and unregistered users.
	// 2: Unregistered users only.
	Unregged int
}


// Command dispatcher.
// Secretly just a pretty map, but don't tell anyone.
type CommandDispatcher map[string]*Command


// NewCommandDispatcher returns a new command dispatcher.
func NewCommandDispatcher() (d CommandDispatcher) {
	d = make(map[string]*Command)
	return
}

// Add adds a command to the dispatcher.
// Takes a name and command structure. The name MUST be uppercase.
// While multiple concurrent Lookup calls are permitted, the package using the
// dispatcher must guarantee that Add is not called at the same time as any.
func (d CommandDispatcher) Add(name string, c *Command) {
	d[name] = c
}

// Lookup looks up whether a command exists in the dispatcher, and returns it.
// If a command does not exist, returns nil. Case-insensitive.
func (d CommandDispatcher) Lookup(name string) (c *Command) {
	return d[strings.ToUpper(name)]
}
