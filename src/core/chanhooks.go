package core

var hookChanAdd *hook
var hookChanDataChanges *hook
var hookChanUserJoin *hook
var hookChanUserRemove *hook
var hookChanDelete *hook

var hookChanDataChange map[string]*hook
var hookChanMessage map[string]*hook


func init() {
	hookChanDataChange = make(map[string]*hook)
	hookChanMessage = make(map[string]*hook)
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

// HookChanUserJoin adds a hook on a user joining a channel.
// The hook receives the user and channel.
// It is illegal to remove the user from the channel in response to this hook.
// It's also stupid- use a permissions hook and deny them joining.
func HookChanUserJoin(f func(*User, *Channel)) {
	h := new(hook)
	h.f = f
	h.next = hookChanUserJoin
	hookChanUserJoin = h
}

// HookChanUserRemove adds a hook on a user being removed from a channel.
// The hook receives the source of the removal, the user, and the channel.
// It must be prepared for source to be nil.
func HookChanUserRemove(f func(*User, *User, *Channel)) {
	h := new(hook)
	h.f = f
	h.next = hookChanUserRemove
	hookChanUserRemove = h
}

// HookChanDelete adds a hook called whenever a channel is deleted.
func HookChanDelete(f func(*Channel)) {
	h := new(hook)
	h.f = f
	h.next = hookChanDelete
	hookChanDelete = h
}

// HookChanMessage adds a hook called whenever a channel receives a message.
// t indicates the type of PM the hook is interested in, and may be "", to
// hook the default type.
// The hook receives the source, target, and message as parameters, and must be
// prepared for the source to be nil.
func HookChanMessage(t string, f func(*User, *Channel, []byte)) {
	h := new(hook)
	h.f = f
	h.next = hookChanMessage[t]
	hookChanMessage[t] = h
}


func runChanAddHooks(ch *Channel) {
	for h := hookChanAdd; h != nil; h = h.next {
		if f := h.f.(func(*Channel)); f != nil {
			f(ch)
		}
	}
}

func runChanUserJoinHooks(u *User, ch *Channel) {
	for h := hookChanUserJoin; h != nil; h = h.next {
		if f := h.f.(func(*User, *Channel)); f != nil {
			f(u, ch)
		}
	}
}

func runChanUserRemoveHooks(source *User, u *User, ch *Channel) {
	for h := hookChanUserRemove; h != nil; h = h.next {
		if f := h.f.(func(*User, *User, *Channel)); f != nil {
			f(source, u, ch)
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

func runChanMessageHooks(source *User, ch *Channel, message []byte, t string) {
	for h := hookChanMessage[t]; h != nil; h = h.next {
		if f := h.f.(func(*User, *Channel, []byte)); f != nil {
			f(source, ch, message)
		}
	}
}
