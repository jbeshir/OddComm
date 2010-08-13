package core

var hookChanAdd map[string]*hook
var hookChanDataChanges map[string]*hook
var hookChanUserJoin map[string]*hook
var hookChanUserRemove map[string]*hook
var hookChanDelete map[string]*hook

var hookChanDataChange map[string]map[string]*hook
var hookChanMessage map[string]map[string]*hook


func init() {
	hookChanAdd = make(map[string]*hook)
	hookChanDataChanges = make(map[string]*hook)
	hookChanUserJoin = make(map[string]*hook)
	hookChanUserRemove = make(map[string]*hook)
	hookChanDelete = make(map[string]*hook)

	hookChanDataChange = make(map[string]map[string]*hook)
	hookChanMessage = make(map[string]map[string]*hook)
}


// HookChanAdd adds a hook called whenever a new channel is created.
// t is the type of channel to hook. "" is default.
func HookChanAdd(t string, f func(*Channel)) {
	h := new(hook)
	h.f = f
	h.next = hookChanAdd[t]
	hookChanAdd[t] = h
}

// HookChanDataChange adds a hook called whenever a channel's metadata changes.
// t is the type of channel to hook. "" is default.
// The name is the metadata whose changes the hook wishes to receive.
// The hook receives the source and target of the change, and the old and new
// values of the data as parameters, and must be prepared for source to be nil.
// "" means unset, for either the old or new value.
func HookChanDataChange(t, name string, f func(*User, *Channel, string, string)) {
	if hookChanDataChange[t] == nil {
		hookChanDataChange[t] = make(map[string]*hook)
	}
	h := new(hook)
	h.f = f
	h.next = hookChanDataChange[t][name]
	hookChanDataChange[t][name] = h
}

// HookChanDataChanges adds a hook called for all channel metadata changes.
// t is the type of channel to hook. "" is default.
// The hook receives the source and target of the change, and lists of
// DataChanges and OldData as parameters, so multiple changes at once result
// in a single call. It must be prepared for source to be nil.
func HookChanDataChanges(t string, f func(*User, *Channel, *DataChange, *OldData)) {
	h := new(hook)
	h.f = f
	h.next = hookChanDataChanges[t]
	hookChanDataChanges[t] = h
}

// HookChanUserJoin adds a hook on a user joining a channel.
// t is the type of channel to hook. "" is default.
// The hook receives the user and channel.
// It is illegal to remove the user from the channel in response to this hook.
// It's also stupid- use a permissions hook and deny them joining.
func HookChanUserJoin(t string, f func(*User, *Channel)) {
	h := new(hook)
	h.f = f
	h.next = hookChanUserJoin[t]
	hookChanUserJoin[t] = h
}

// HookChanUserRemove adds a hook on a user being removed from a channel.
// t is the type of channel to hook. "" is default.
// The hook receives the source of the removal, the user, and the channel.
// It must be prepared for source to be nil.
func HookChanUserRemove(t string, f func(*User, *User, *Channel)) {
	h := new(hook)
	h.f = f
	h.next = hookChanUserRemove[t]
	hookChanUserRemove[t] = h
}

// HookChanDelete adds a hook called whenever a channel is deleted.
// t is the type of channel to hook. "" is default.
func HookChanDelete(t string, f func(*Channel)) {
	h := new(hook)
	h.f = f
	h.next = hookChanDelete[t]
	hookChanDelete[t] = h
}

// HookChanMessage adds a hook called whenever a channel receives a message.
// chantype indicates the type of channel, and msgtype indicates the type of
// message the hook is interested in, and either may be "" for the default.
// The hook receives the source, target, and message as parameters, and must be
// prepared for the source to be nil.
func HookChanMessage(chantype, msgtype string, f func(*User, *Channel, []byte)) {
	if hookChanMessage[chantype] == nil {
		hookChanMessage[chantype] = make(map[string]*hook)
	}
	h := new(hook)
	h.f = f
	h.next = hookChanMessage[chantype][msgtype]
	hookChanMessage[chantype][msgtype] = h
}


func runChanAddHooks(t string, ch *Channel) {
	for h := hookChanAdd[t]; h != nil; h = h.next {
		if f := h.f.(func(*Channel)); f != nil {
			f(ch)
		}
	}
}

func runChanUserJoinHooks(t string, u *User, ch *Channel) {
	for h := hookChanUserJoin[t]; h != nil; h = h.next {
		if f := h.f.(func(*User, *Channel)); f != nil {
			f(u, ch)
		}
	}
}

func runChanUserRemoveHooks(t string, source *User, u *User, ch *Channel) {
	for h := hookChanUserRemove[t]; h != nil; h = h.next {
		if f := h.f.(func(*User, *User, *Channel)); f != nil {
			f(source, u, ch)
		}
	}
}

func runChanDataChangeHooks(t string, source *User, ch *Channel, name, oldvalue, newvalue string) {
	if hookChanDataChange[t] == nil {
		return
	}
	for h := hookChanDataChange[t][name]; h != nil; h = h.next {
		if f := h.f.(func(*User, *Channel, string, string)); f != nil {
			f(source, ch, oldvalue, newvalue)
		}
	}
}

func runChanDataChangesHooks(t string, source *User, ch *Channel, c *DataChange, o *OldData) {
	for h := hookChanDataChanges[t]; h != nil; h = h.next {
		if f := h.f.(func(*User, *Channel, *DataChange, *OldData)); f != nil {
			f(source, ch, c, o)
		}
	}
}

func runChanMessageHooks(t, msgt string, source *User, ch *Channel, message []byte) {
	if hookChanMessage[t] == nil {
		return
	}
	for h := hookChanMessage[t][msgt]; h != nil; h = h.next {
		if f := h.f.(func(*User, *Channel, []byte)); f != nil {
			f(source, ch, message)
		}
	}
}
