package core

var hookMemberDataChange = make(map[string]map[string][]func(string, *User, *Membership, string, string))


// HookMemberDataChange adds a hook called whenever a channel membership's
// metadata changes.
// t is the type of channel to hook. "" is default.
// The name is the metadata whose changes the hook wishes to receive.
// The hook receives the source and target of the change, and the old and new
// values of the data as parameters, and must be prepared for source to be nil.
// "" means unset, for either the old or new value.
func HookMemberDataChange(t, name string, f func(string, *User, *Membership, string, string)) {
	if hookMemberDataChange[t] == nil {
		hookMemberDataChange[t] = make(map[string][]func(string, *User, *Membership, string, string))
	}
	hookMemberDataChange[t][name] = append(hookMemberDataChange[t][name], f)
}


func runMemberDataChangeHooks(pkg, t string, source *User, m *Membership, name, oldvalue, newvalue string) {
	if hookMemberDataChange[t] == nil {
		return
	}
	for _, f := range hookMemberDataChange[t][name] {
		f(pkg, source, m, oldvalue, newvalue)
	}
}
