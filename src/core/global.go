package core

import "sync"

import "oddcomm/lib/trie"


// Global stores global (meta)data, for server-wide data storage.
var global = new(globalData)
var Global Extensible = global

// Define a type wrapping our global data trie.
// This lets us provide it with methods to meet the Extensible interface.
type globalData struct {
	data trie.StringTrie
	mutex sync.Mutex
}


// SetData sets the given single piece of global data.
// source may be nil, in which case the metadata is set by the server.
// Setting it to "" unsets it.
func (g *globalData) SetData(source *User, name, value string) {
	var oldvalue string

	g.mutex.Lock()

	if value != "" {
		oldvalue = g.data.Insert(name, value)
	} else {
		oldvalue = g.data.Remove(name)
	}

	// If nothing changed, don't call hooks.
	if oldvalue != value {
		var c DataChange
		c.Name = name
		c.Data = value

		hookRunner <- func() {
			runGlobalDataChangeHooks(source, name, oldvalue, value)
			runGlobalDataChangesHooks(source, []DataChange{c}, []string{oldvalue})
		}
	}

	g.mutex.Unlock()
}


// SetDataList performs the given list of global data changes.
// This is equivalent to lots of SetData calls, except hooks for all data
// changes will receive it as a single list, and it is cheaper.
// source may be nil, in which case the metadata is set by the server.
func (g *globalData) SetDataList(source *User, changes []DataChange) {
	done := make([]DataChange, 0, len(changes))
	old := make([]string, 0, len(changes))

	g.mutex.Lock()

	for _, it := range changes {

		// Make the change.
		var oldvalue string
		if it.Data != "" {
			oldvalue = g.data.Insert(it.Name, it.Data)
		} else {
			oldvalue = g.data.Remove(it.Name)
		}

		// If this was a do-nothing change, don't report it.
		if oldvalue == it.Data {
			continue
		}

		// Otherwise, add to be sent to hooks.
		done = append(done, it)
		old = append(old, oldvalue)
	}

	hookRunner <- func() {
		for i, it := range done {
			runGlobalDataChangeHooks(source, it.Name, old[i], it.Data)
		}
		runGlobalDataChangesHooks(source, done, old)
	}

	g.mutex.Unlock()
}


// Data gets the given piece of global data.
// If it is not set, this method returns "".
func (g *globalData) Data(name string) (value string) {
	return g.data.Get(name)
}

// DataRange calls the given function for every piece of global data with the
// given prefix. If none are found, the function is never called. Items added
// while this function is running may or may not be missed.
func (g *globalData) DataRange(prefix string, f func(name, value string)) {
	for it := g.data.GetSub(prefix); it != nil; {
		name, data := it.Value()
		f(name, data)
		if !it.Next() {
			it = nil
		}
	}
}
