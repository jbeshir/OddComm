package core

// Represents anything which has Data and SetData methods for metadata.
type Extensible interface {
	Data(name string) string
	SetData(name string, value string)
}

// Represents a metadata change.
// The name is the name of the metadata changed, and the data is what it is
// set to.
type DataChange struct {
	Name, Data string
	Next *DataChange
}

// Represents a previous metadata value.
// In conjunction with a DataChange list, permits viewing the previous value
// of a changed metadata item.
type OldData struct {
	Data string
	Next *OldData
}

// Represents a hook.
type hook struct {
	next *hook
	f interface{}
}


var hookStart *hook


// HookStart adds a hook called on server startup.
// This is called after starting the server subsystems.
func HookStart(f func()) {
	h := new(hook)
	h.f = f
	h.next = hookStart
	hookStart = h
}


// RunStartHooks calls all hooks added with HookStart. This should only be
// called once, and by main.
func RunStartHooks() {
	for h := hookStart; h != nil; h = h.next {
		if f, ok := h.f.(func()); ok && f != nil {
			f()
		}
	}
}
