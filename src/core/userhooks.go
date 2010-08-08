package core

var holdRegistration map[string]int

// Hook list for user hooks.
// Contains sublists for all users and just regged users.
type userHooklist struct {
	all *hook
	regged *hook
}

var hookUserAdd userHooklist
var hookUserRegister userHooklist
var hookUserNickChange userHooklist
var hookUserDataChanges userHooklist
var hookUserRemoved userHooklist

var hookUserDataChange map[string]*userHooklist
var hookUserPM map[string]*userHooklist


func init() {
	holdRegistration = make(map[string]int)
	hookUserDataChange = make(map[string]*userHooklist)
	hookUserPM = make(map[string]*userHooklist)
}


// RegistrationHold causes user registration for new users to be held until the
// package which called it calls PermitRegistration on the user.
// Packages wishing to filter new users, or perform operations on them prior to
// them connecting, should call this.
// A non-nil creator specifies the creating package whose users are to be held.
// This should only be used by a subsystem or module to hold its own users.
// A nil creator holds all new users not marked as pre-checked.
func RegistrationHold(creator string) {
	holdRegistration[creator]++
}


// Add a hook to a user hook list.
func (l *userHooklist) add(f interface{}, unregged bool) {
	h := new(hook)
	h.f = f

	var list **hook
	if unregged {
		list = &l.all
	} else {
		list = &l.regged
	}

	for *list != nil {
		list = &((*list).next)
	}
	*list = h
}

// Run all the hooks on a user hook list.
func (l *userHooklist) run(f func(interface{}), registered bool) {
	for h := l.all; h != nil; h = h.next {
		f(h.f)
	}

	if !registered { return }
	
	for h := l.regged; h != nil; h = h.next {
		f(h.f)
	}
}


// HookUserAdd adds a hook called whenever a new user is added.
// The hook receives the name of the creating module as a parameter.
func HookUserAdd(f func(*User, string)) {
	hookUserAdd.add(f, true)
}

// HookUserRegister adds a hook called whenever a user completes registration.
func HookUserRegister(f func(*User)) {
	hookUserRegister.add(f, false)
}

// HookUserNickChange adds a hook called whenever a user changes nick.
// If unregged is false, it is not called for unregistered users.
// The hook receives the old and new nicks as parameters.
func HookUserNickChange(f func(*User, string, string), unregged bool) {
	hookUserNickChange.add(f, unregged)
}

// HookUserDataChange adds a hook called whenever a user's metadata changes.
// The name is the metadata whose changes the hook wishes to receive.
// If unregged is false, it is not called for unregistered users.
// The hook receives the old and new values of the data as parameters.
// "" means unset.
func HookUserDataChange(name string, f func(*User, string, string),
                        unregged bool) {
	if hookUserDataChange[name] == nil {
		hookUserDataChange[name] = new(userHooklist)
	}

	hookUserDataChange[name].add(f, unregged)		
}

// HookUserDataChanges adds a hook called for all user metadata changes.
// If unregged is false, it is not called for unregistered users.
// The hook receives a list of UserDataChanges as a parameter, so multiple
// changes at once result in a single call. It does not have access to the
// previous values of those metadata entries.
func HookUserDataChanges(f func(*User, *UserDataChange), unregged bool) {
	hookUserDataChanges.add(f, unregged)
}


// HookUserPM adds a hook called whenever a user sends/receives a PM.
// t indicates the type of PM the hook is interested in, and may be "", to
// hook the default type.
// The hook receives the source, target, and message as parameters, and must be
// prepared for the source to be nil.
func HookUserPM(t string, f func(*User, *User, []byte)) {
	if hookUserPM[t] == nil {
		hookUserPM[t] = new(userHooklist)
	}

	hookUserPM[t].add(f, false)
}

// HookUserRemoved adds a hook called whenever a user is removed.
// If unregged is false, it is not called for unregistered users.
func HookUserRemoved(f func(*User, string), unregged bool) {
	hookUserRemoved.add(f, unregged)
}


func runUserAddHooks(u *User, creator string) {
	hookUserAdd.run(func(f interface{}) {
		if h := f.(func(*User, string)); h != nil {
			h(u, creator)
		}
	}, false)
}

func runUserRegisterHooks(u *User) {
	hookUserRegister.run(func(f interface{}) {
		if h, ok := f.(func(*User)); ok && h != nil {
			h(u)
		}
	}, true)
}

func runUserNickChangeHooks(u *User, oldnick, newnick string) {
	hookUserNickChange.run(func(f interface{}) {
		if h, ok := f.(func(*User, string, string)); ok && h != nil {
			h(u, oldnick, newnick)
		}
	}, u.Registered())
}

func runUserDataChangesHooks(u *User, changes *UserDataChange) {
	hookUserDataChanges.run(func(f interface{}) {
		if h, ok := f.(func(*User, *UserDataChange)); ok && h != nil {
			h(u, changes)
		}
	}, u.Registered())
}

func runUserRemovedHooks(u *User, message string) {
	hookUserRemoved.run(func(f interface{}) {
		if h, ok := f.(func(*User, string)); ok && h != nil {
			h(u, message)
		}
	}, u.Registered())
}

func runUserDataChangeHooks(u *User, name string, oldvalue, newvalue string) {
	if hookUserDataChange[name] == nil { return }
	hookUserDataChange[name].run(func(f interface{}) {
		if h, ok := f.(func(*User, string, string)); ok && h != nil {
			h(u, oldvalue, newvalue)
		}
	}, u.Registered())
}

func runUserPMHooks(source, target *User, message []byte, t string) {
	if hookUserPM[t] == nil { return }
	hookUserPM[t].run(func(f interface{}) {
		if h, ok := f.(func(*User, *User, []byte)); ok && h != nil {
			h(source, target, message)
		}
	}, true)
}
