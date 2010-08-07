package core

var holdRegistration map[string]int

var hookUserAdd hooklist
var hookUserRegister hooklist
var hookUserNickChange hooklist
var hookUserRemoved hooklist

var hookUserDataChange map[string]*hooklist


func init() {
	holdRegistration = make(map[string]int)
	hookUserDataChange = make(map[string]*hooklist)
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

// HookUserRemoved adds a hook called whenever a user is removed.
// If unregged is false, it is not called for unregistered users.
func HookUserRemoved(f func(*User, string), unregged bool) {
	hookUserRemoved.add(f, unregged)
}

// HookUserDataChange adds a hook called whenever a user's metadata changes.
// The name is the metadata whose changes the hook wishes to receive.
// If unregged is false, it is not called for unregistered users.
// The hook receives the old and new values of the data as parameters.
// "" means unset.
func HookUserDataChange(name string, f func(*User, string, string),
                        unregged bool) {
	if hookUserDataChange[name] == nil {
		hookUserDataChange[name] = new(hooklist)
	}

	hookUserDataChange[name].add(f, unregged)		
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
