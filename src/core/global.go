package core

import "sync"

import "oddcomm/lib/trie"


// Global stores global (meta)data, for server-wide data storage.
var Global Extensible = new(globalData)
var globalMutex sync.Mutex

// Define a type wrapping our global data trie.
// This lets us provide it with methods to meet the Extensible interface.
type globalData struct {
	data trie.StringTrie
}


// SetData sets the given single piece of global data.
// source may be nil, in which case the metadata is set by the server.
// Setting it to "" unsets it.
func (g *globalData) SetData(source *User, name, value string) {
	var oldvalue string

	globalMutex.Lock()
	if value != "" {
		oldvalue = g.data.Add(name, value)
	} else {
		oldvalue = g.data.Del(name)
	}
	globalMutex.Unlock()

	// If nothing changed, don't call hooks.
	if oldvalue == value {
		return
	}

	c := new(DataChange)
	c.Name = name
	c.Data = value
	old := new(OldData)
	old.Data = value
	runGlobalDataChangeHooks(source, name, oldvalue, value)
	runGlobalDataChangesHooks(source, c, old)
}


// SetDataList performs the given list of global data changes.
// This is equivalent to lots of SetData calls, except hooks for all data
// changes will receive it as a single list, and it is cheaper.
// source may be nil, in which case the metadata is set by the server.
func (g *globalData) SetDataList(source *User, c *DataChange) {
	var oldvalues *OldData
	var lasthook *DataChange
	var lastold **OldData = &oldvalues

	globalMutex.Lock()
	for it := c; it != nil; it = it.Next {

		// Make the change.
		var oldvalue string
		if it.Data != "" {
			oldvalue = g.data.Add(it.Name, it.Data)
		} else {
			oldvalue = g.data.Del(it.Name)
		}

		// If this was a do-nothing change, cut it out.
		if oldvalue == it.Data {
			if lasthook != nil {
				lasthook.Next = it.Next
			} else {
				c = it.Next
			}
			continue
		}

		olddata := new(OldData)
		olddata.Data = oldvalue
		*lastold = olddata
		lasthook = it
		lastold = &olddata.Next
	}
	globalMutex.Unlock()

	for it, old := c, oldvalues; it != nil && old != nil; it, old = it.Next, old.Next {
		runGlobalDataChangeHooks(source, c.Name, old.Data, c.Data)
	}
	runGlobalDataChangesHooks(source, c, oldvalues)
}


// Data gets the given piece of global data.
// If it is not set, this method returns "".
func (g *globalData) Data(name string) (value string) {
	globalMutex.Lock()
	value = g.data.Get(name)
	globalMutex.Unlock()

	return
}

// DataRange calls the given function for every piece of global data with the
// given prefix. If none are found, the function is never called. Items added
// while this function is running may or may not be missed.
func (g *globalData) DataRange(prefix string, f func(name, value string)) {
	var dataArray [100]DataChange
	var data []DataChange = dataArray[0:0]
	var root, it trie.StringTrie

	for firstrun := true; firstrun || !it.Empty(); {

		globalMutex.Lock()

		// On first run, get an iterator pointing to our first value.
		if firstrun {
			root = g.data.GetSub(prefix)
			it = root
			if !it.Empty() {
				if key, _ := it.Value(); key == "" {
					it = it.Next(root)
				}
			}
			firstrun = false
		}

		// Get up to 100 values from this subtrie.
		for i := 0; !it.Empty() && i < cap(data); i++ {
			data = data[0 : i+1]
			data[i].Name, data[i].Data = it.Value()
			it = it.Next(root)
		}

		globalMutex.Unlock()

		// Call the function for all of them, and clear data.
		for _, item := range data {
			f(item.Name, item.Data)
		}
		data = data[0:0]
	}
}
