package core

// Global stores global (meta)data, for server-wide data storage.
var Global Extensible = new(globalData)

// Define a type wrapping our global data trie.
// This lets us provide it with methods to meet the Extensible interface.
type globalData struct {
	data *Trie
}


// SetData sets the given single piece of global data.
// source may be nil, in which case the metadata is set by the server.
// Setting it to "" unsets it.
func (g *globalData) SetData(source *User, name, value string) {
	var oldvalue string
	
	wait := make(chan bool)
	corechan <- func() {
		var old interface{}
		if value != "" {
			old = TrieAdd(&g.data, name, value)
		} else {
			old = TrieDel(&g.data, name)
		}
		if old != nil {
			oldvalue = old.(string)
		}

		wait <- true
	}
	<-wait
	
	// If nothing changed, don't call hooks.
	if oldvalue == value {
		return
	}

	runGlobalDataChangeHooks(source, name, oldvalue, value)

	c := new(DataChange)
	c.Name = name
	c.Data = value
	old := new(OldData)
	old.Data = value
	runGlobalDataChangesHooks(source, c, old)
}


// SetDataList performs the given list of global data changes.
// This is equivalent to lots of SetData calls, except hooks for all data
// changes will receive it as a single list, and it is cheaper.
// source may be nil, in which case the metadata is set by the server.
func (g *globalData) SetDataList(source *User, c *DataChange) {
	var oldvalues *OldData
	wait := make(chan bool)
	corechan <- func() {
		var lasthook *DataChange
		var lastold **OldData = &oldvalues
		for it := c; it != nil; it = it.Next {

			// Make the change.
			var old interface{}
			var oldvalue string
			if it.Data != "" {
				old = TrieAdd(&g.data, it.Name, it.Data)
			} else {
				old = TrieDel(&g.data, it.Name)
			}
			if old != nil {
				oldvalue = old.(string)
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

		wait <- true
	}
	<-wait

	for it, old := c, oldvalues; it != nil && old != nil; it, old = it.Next, old.Next {
		runGlobalDataChangeHooks(source, c.Name, old.Data, c.Data)
	}
	runGlobalDataChangesHooks(source, c, oldvalues)
}


// Data gets the given piece of global data.
// If it is not set, this method returns "".
func (g *globalData) Data(name string) (value string) {
	wait := make(chan bool)
	corechan <- func() {
		val := TrieGet(&g.data, name)
		if val != nil {
			value = val.(string)
		}
		wait <- true
	}
	<-wait

	return
}

// DataRange calls the given function for every piece of global data with the
// given prefix. If none are found, the function is never called. Items added
// while this function is running may or may not be missed.
func (g *globalData) DataRange(prefix string, f func(name, value string)) {
	var dataArray [50]DataChange
	var data []DataChange = dataArray[0:0]
	var root, it *Trie
	wait := make(chan bool)

	// Get an iterator pointing to our first value.
	corechan <- func() {
		root = TrieGetSub(&g.data, prefix)
		it = root
		if it != nil {
			if key, _ := it.Value(); key == "" {
				it = it.Next(root)
			}
		}
		wait <- true
	}
	<-wait

	for it != nil {
		// Get up to 50 values from this subtrie.
		corechan <- func() {
			for i := 0; i < cap(data); i++ {
				var val interface{}
				data = data[0:i+1]
				data[i].Name, val = it.Value()
				data[i].Data = val.(string)
				it = it.Next(root)
				if it == nil {
					break
				}
			}
			wait <- true
		}
		<- wait

		// Call the function for all of them, and clear data.
		for _, item := range data {
			f(item.Name, item.Data)
		}
		data = data[0:0]
	}
}
