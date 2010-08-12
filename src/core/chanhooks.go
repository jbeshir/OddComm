package core

var hookChanAdd *hook
var hookChanDataChanges *hook
var hookChanDelete *hook

var hookChanDataChange map[string]*hook


func init() {
	hookChanDataChange = make(map[string]*hook)
}


// HookChanAdd adds a hook called whenever a new channel is created.
func HookChanAdd(f func(*Channel)) {
	h := new(hook)
	h.f = f
	h.next = hookChanAdd
	hookChanAdd = h
}

// HookChanDataChange adds a hook called whenever a channel's metadata changes.
// The name is the metadata whose changes the hook wishes to receive.
// The hook receives the source and target of the change, and the old and new
// values of the data as parameters, and must be prepared for source to be nil.
// "" means unset, for either the old or new value.
func HookChanDataChange(name string, f func(*User, *Channel, string, string)) {
	h := new(hook)
	h.f = f
	h.next = hookChanDataChange[name]
	hookChanDataChange[name] = h
}

// HookChanDataChanges adds a hook called for all channel metadata changes.
// The hook receives the source and target of the change, and lists of
// DataChanges and OldData as parameters, so multiple changes at once result
// in a single call. It must be prepared for source to be nil.
func HookChanDataChanges(f func(*User, *Channel, *DataChange, *OldData)) {
	h := new(hook)
	h.f = f
	h.next = hookChanDataChanges
	hookChanDataChanges = h
}

// HookChanDelete adds a hook called whenever a channel is deleted.
func HookChanDelete(f func(*Channel)) {
	h := new(hook)
	h.f = f
	h.next = hookChanDelete
	hookChanDelete = h
}


func runChanAddHooks(ch *Channel) {
	for h := hookChanAdd; h != nil; h = h.next {
		if f := h.f.(func(*Channel)); f != nil {
			f(ch)
		}
	}
}

func runChanDataChangeHooks(source *User, ch *Channel, name, oldvalue, newvalue string) {
	for h := hookChanDataChange[name]; h != nil; h = h.next {
		if f := h.f.(func(*User, *Channel, string, string)); f != nil {
			f(source, ch, oldvalue, newvalue)
		}
	}
}

func runChanDataChangesHooks(source *User, ch *Channel, c *DataChange, o *OldData) {
	for h := hookChanDataChanges; h != nil; h = h.next {
		if f := h.f.(func(*User, *Channel, *DataChange, *OldData)); f != nil {
			f(source, ch, c, o)
		}
	}
}
