package core

var hookGlobalDataChanges []func(*User, *DataChange, *OldData)
var hookGlobalDataChange = make(map[string][]func(*User, string, string))


// HookGlobalDataChange adds a hook called whenever global data changes.
// The name is the data whose changes the hook wishes to receive.
// The hook receives the source of the change, and the old and new values of
// the data as parameters, and must be prepared for source to be nil.
// "" means unset, for either the old or new value.
func HookGlobalDataChange(name string, f func(*User, string, string)) {
	hookGlobalDataChange[name] = append(hookGlobalDataChange[name], f)
}

// HookGlobalDataChanges adds a hook called for all global data changes.
// The hook receives the source of the change, and lists of DataChanges and
// OldData as parameters, so multiple changes at once result in a single call.
// It must be prepared for source to be nil.
func HookGlobalDataChanges(f func(*User, *DataChange, *OldData)) {
	hookGlobalDataChanges = append(hookGlobalDataChanges, f)
}

func runGlobalDataChangeHooks(source *User, name, oldvalue, newvalue string) {
	for _, f := range hookGlobalDataChange[name] {
		f(source, oldvalue, newvalue)
	}
}

func runGlobalDataChangesHooks(source *User, c *DataChange, o *OldData) {
	for _, f := range hookGlobalDataChanges {
		f(source, c, o)
	}
}
