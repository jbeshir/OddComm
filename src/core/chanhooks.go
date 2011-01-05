package core

var hookChanAdd = make(map[string][]func(*Channel))
var hookChanDataChanges = make(map[string][]func(*User, *Channel, *DataChange, *OldData))
var hookChanUserJoin = make(map[string][]func(*Channel, []*User))
var hookChanUserRemove = make(map[string][]func(*User, *User, *Channel, string))
var hookChanDelete = make(map[string][]func(*Channel))

var hookChanDataChange = make(map[string]map[string][]func(*User, *Channel, string, string))
var hookChanMessage = make(map[string]map[string][]func(*User, *Channel, []byte))


// HookChanAdd adds a hook called whenever a new channel is created.
// t is the type of channel to hook. "" is default.
func HookChanAdd(t string, f func(*Channel)) {
	hookChanAdd[t] = append(hookChanAdd[t], f)
}

// HookChanDataChange adds a hook called whenever a channel's metadata changes.
// t is the type of channel to hook. "" is default.
// The name is the metadata whose changes the hook wishes to receive.
// The hook receives the source and target of the change, and the old and new
// values of the data as parameters, and must be prepared for source to be nil.
// "" means unset, for either the old or new value.
func HookChanDataChange(t, name string, f func(*User, *Channel, string, string)) {
	if hookChanDataChange[t] == nil {
		hookChanDataChange[t] = make(map[string][]func(*User, *Channel, string, string))
	}
	hookChanDataChange[t][name] = append(hookChanDataChange[t][name], f)
}

// HookChanDataChanges adds a hook called for all channel metadata changes.
// t is the type of channel to hook. "" is default.
// The hook receives the source and target of the change, and lists of
// DataChanges and OldData as parameters, so multiple changes at once result
// in a single call. It must be prepared for source to be nil.
func HookChanDataChanges(t string, f func(*User, *Channel, *DataChange, *OldData)) {
	hookChanDataChanges[t] = append(hookChanDataChanges[t], f)
}

// HookChanUserJoin adds a hook on a user joining a channel.
// t is the type of channel to hook. "" is default.
// The hook receives the user and channel.
// It is illegal to remove the user from the channel in response to this hook.
// It's also stupid- use a permissions hook and deny them joining.
func HookChanUserJoin(t string, f func(*Channel, []*User)) {
	hookChanUserJoin[t] = append(hookChanUserJoin[t], f)
}

// HookChanUserRemove adds a hook on a user being removed from a channel.
// t is the type of channel to hook. "" is default.
// The hook receives the source of the removal, the user, the channel, and the
// message associated with the removal.
// It must be prepared for source to be nil.
func HookChanUserRemove(t string, f func(*User, *User, *Channel, string)) {
	hookChanUserRemove[t] = append(hookChanUserRemove[t], f)
}

// HookChanDelete adds a hook called whenever a channel is deleted.
// t is the type of channel to hook. "" is default.
func HookChanDelete(t string, f func(*Channel)) {
	hookChanDelete[t] = append(hookChanDelete[t], f)
}

// HookChanMessage adds a hook called whenever a channel receives a message.
// chantype indicates the type of channel, and msgtype indicates the type of
// message the hook is interested in, and either may be "" for the default.
// The hook receives the source, target, and message as parameters, and must be
// prepared for the source to be nil.
func HookChanMessage(chant, msgt string, f func(*User, *Channel, []byte)) {
	if hookChanMessage[chant] == nil {
		hookChanMessage[chant] = make(map[string][]func(*User, *Channel, []byte))
	}
	hookChanMessage[chant][msgt] = append(hookChanMessage[chant][msgt], f)
}


func runChanAddHooks(t string, ch *Channel) {
	for _, f := range hookChanAdd[t] {
		f(ch)
	}
}

func runChanUserJoinHooks(t string, ch *Channel, users []*User) {
	for _, f := range hookChanUserJoin[t] {
		f(ch, users)
	}
}

func runChanUserRemoveHooks(t string, source *User, target *User, ch *Channel, message string) {
	for _, f := range hookChanUserRemove[t] {
		f(source, target, ch, message)
	}
}

func runChanDataChangeHooks(t string, source *User, ch *Channel, name, oldvalue, newvalue string) {
	if hookChanDataChange[t] == nil {
		return
	}
	for _, f := range hookChanDataChange[t][name] {
		f(source, ch, oldvalue, newvalue)
	}
}

func runChanDataChangesHooks(t string, source *User, ch *Channel, c *DataChange, o *OldData) {
	for _, f := range hookChanDataChanges[t] {
		f(source, ch, c, o)
	}
}

func runChanMessageHooks(t, msgt string, source *User, ch *Channel, message []byte) {
	if hookChanMessage[t] == nil {
		return
	}
	for _, f := range hookChanMessage[t][msgt] {
		f(source, ch, message)
	}
}
