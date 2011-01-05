package core

// Represents anything which has Data and SetData methods for metadata.
type Extensible interface {
	Data(name string) string
	SetData(source *User, name, value string)
	SetDataList(source *User, c *DataChange)
	DataRange(prefix string, f func(name, value string))
}

// Represents a metadata change.
// The name is the name of the metadata changed, and the data is what it is
// set to. The member value, only valid for channel metadata changes, is nil
// for non-membership changes, or the Membership structure which was changed
// changed for membership changes.
type DataChange struct {
	Name, Data string
	Member     *Membership
	Next       *DataChange
}

// Represents a previous metadata value.
// In conjunction with a DataChange list, permits viewing the previous value
// of a changed metadata item.
type OldData struct {
	Data string
	Next *OldData
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
