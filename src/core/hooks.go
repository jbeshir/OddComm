package core


// Represents a metadata change.
// The name is the name of the metadata changed, and the data is what it is
// set to.
type DataChange struct {
	Name, Data string
	Next *DataChange
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
