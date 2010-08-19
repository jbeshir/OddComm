package core

var hookMemberDataChange map[string]map[string]*hook

func init() {
	hookMemberDataChange = make(map[string]map[string]*hook)
}


// HookMemberDataChange adds a hook called whenever a channel membership's
// metadata changes.
// t is the type of channel to hook. "" is default.
// The name is the metadata whose changes the hook wishes to receive.
// The hook receives the source and target of the change, and the old and new
// values of the data as parameters, and must be prepared for source to be nil.
// "" means unset, for either the old or new value.
func HookMemberDataChange(t, name string, f func(*User, *Membership, string, string)) {
	if hookMemberDataChange[t] == nil {
		hookMemberDataChange[t] = make(map[string]*hook)
	}
	h := new(hook)
	h.f = f
	h.next = hookMemberDataChange[t][name]
	hookMemberDataChange[t][name] = h
}


func runMemberDataChangeHooks(t string, source *User, m *Membership, name, oldvalue, newvalue string) {
	if hookMemberDataChange[t] == nil {
		return
	}
	for h := hookMemberDataChange[t][name]; h != nil; h = h.next {
		if f := h.f.(func(*User, *Membership, string, string)); f != nil {
			f(source, m, oldvalue, newvalue)
		}
	}
}

