package core

var hookGlobalDataChanges *hook
var hookGlobalDataChange map[string]*hook


func init() {
	hookGlobalDataChange = make(map[string]*hook)
}


// HookGlobalDataChange adds a hook called whenever global data changes.
// The name is the data whose changes the hook wishes to receive.
// The hook receives the source of the change, and the old and new values of
// the data as parameters, and must be prepared for source to be nil.
// "" means unset, for either the old or new value.
func HookGlobalDataChange(name string, f func(*User, string, string)) {
	h := new(hook)
	h.f = f
	h.next = hookGlobalDataChange[name]
	hookGlobalDataChange[name] = h
}

// HookGlobalDataChanges adds a hook called for all global data changes.
// The hook receives the source of the change, and lists of DataChanges and
// OldData as parameters, so multiple changes at once result in a single call.
// It must be prepared for source to be nil.
func HookGlobalDataChanges(f func(*User, *DataChange, *OldData)) {
	h := new(hook)
	h.f = f
	h.next = hookGlobalDataChanges
	hookGlobalDataChanges = h
}

func runGlobalDataChangeHooks(source *User, name, oldvalue, newvalue string) {
	for h := hookGlobalDataChange[name]; h != nil; h = h.next {
		if f := h.f.(func(*User, string, string)); f != nil {
			f(source, oldvalue, newvalue)
		}
	}
}

func runGlobalDataChangesHooks(source *User, c *DataChange, o *OldData) {
	for h := hookGlobalDataChanges; h != nil; h = h.next {
		if f := h.f.(func(*User, *DataChange, *OldData)); f != nil {
			f(source, c, o)
		}
	}
}
