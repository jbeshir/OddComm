package core

// Represents anything which has Data and SetData methods for metadata.
type Extensible interface {
	Data(name string) string
	SetData(source *User, name, value string)
	SetDataList(source *User, changes []DataChange)
	DataRange(prefix string, f func(name, value string))
}

// Represents a metadata change.
// The name is the name of the metadata changed, and the data is what it is
// set to. The member value, only valid for channel metadata changes, is nil
// for non-membership changes, or the Membership structure which was changed
// cfor membership changes.
type DataChange struct {
	Name, Data string
	Member     *Membership
}


var hookStart []func()


// HookStart adds a hook called on server startup.
// This is called after starting the server subsystems.
func HookStart(f func()) {
	hookStart = append(hookStart, f)
}


// RunStartHooks calls all hooks added with HookStart. This should only be
// called once, and by main.
func RunStartHooks() {
	for _, f := range hookStart {
		f()
	}
}
