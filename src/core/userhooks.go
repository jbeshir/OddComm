package core

var holdRegistration = make(map[string]int)


var hookUserAdd []func(string, *User)
var hookUserRegister []func(string, *User)

var hookUserNickChange struct {
	all    []func(string, *User, string, string)
	regged []func(string, *User, string, string)
}

var hookUserDataChanges struct {
	all    []func(string, *User, *User, []DataChange, []string)
	regged []func(string, *User, *User, []DataChange, []string)
}

var hookUserDelete struct {
	all    []func(string, *User, *User, string)
	regged []func(string, *User, *User, string)
}

type hookDataChangeType struct {
        all    []func(string, *User, *User, string, string)
	regged []func(string, *User, *User, string, string)
}
var hookDataChange = make(map[string]hookDataChangeType)

var hookUserMessage = make(map[string][]func(string, *User, *User, []byte))


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
func HookUserAdd(f func(string, *User)) {
	hookUserAdd = append(hookUserAdd, f)
}

// HookUserRegister adds a hook called whenever a user completes registration.
func HookUserRegister(f func(string, *User)) {
	hookUserRegister = append(hookUserRegister, f)
}

// HookUserNickChange adds a hook called whenever a user changes nick.
// If unregged is false, it is not called for unregistered users.
// The hook receives the old and new nicks as parameters.
func HookUserNickChange(f func(string, *User, string, string), unregged bool) {
	if unregged {
		hookUserNickChange.all = append(hookUserNickChange.all, f)
	} else {
		hookUserNickChange.regged = append(hookUserNickChange.regged, f)
	}
}

// HookUserDataChange adds a hook called whenever a user's metadata changes.
// The name is the metadata whose changes the hook wishes to receive.
// If unregged is false, it is not called for unregistered users.
// The hook receives the source and target of the change, and the old and new
// values of the data as parameters, and must be prepared for source to be nil.
// "" means unset, for either the old or new value.
func HookUserDataChange(name string, f func(string, *User, *User, string, string), unregged bool) {
	hooks := hookDataChange[name]
	if unregged {
		hooks.all = append(hooks.all, f)
	} else {
		hooks.regged = append(hooks.regged, f)
	}
	hookDataChange[name] = hooks
}

// HookUserDataChanges adds a hook called for all user metadata changes.
// If unregged is false, it is not called for unregistered users.
// The hook receives the source and target of the change, and lists of
// DataChanges and OldData as parameters, so multiple changes at once result
// in a single call. It must be prepared for source to be nil.
func HookUserDataChanges(f func(string, *User, *User, []DataChange, []string), unregged bool) {
	if unregged {
		hookUserDataChanges.all = append(hookUserDataChanges.all, f)
	} else {
		hookUserDataChanges.regged = append(hookUserDataChanges.regged, f)
	}
}


// HookUserMessage adds a hook called whenever a user receives a message.
// t indicates the type of PM the hook is interested in, and may be "", to
// hook the default type.
// The hook receives the source, target, and message as parameters, and must be
// prepared for the source to be nil.
func HookUserMessage(t string, f func(string, *User, *User, []byte)) {
	hookUserMessage[t] = append(hookUserMessage[t], f)
}

// HookUserDelete adds a hook called whenever a user is deleted.
// The hook receives the source, removed user, and message as parameters, and
// must be prepared for the source to be nil.
// If unregged is false, it is not called for unregistered users.
func HookUserDelete(f func(string, *User, *User, string), unregged bool) {
	if unregged {
		hookUserDelete.all = append(hookUserDelete.all, f)
	} else {
		hookUserDelete.regged = append(hookUserDelete.regged, f)
	}
}


func runUserAddHooks(creator string, u *User) {
	for _, f := range hookUserAdd {
		f(creator, u)
	}
}

func runUserRegisterHooks(pkg string, u *User) {
	for _, f := range hookUserRegister {
		f(pkg, u)
	}
}

func runUserNickChangeHooks(pkg string, u *User, oldnick, newnick string) {
	for _, f := range hookUserNickChange.all {
		f(pkg, u, oldnick, newnick)
	}

	if u.Registered() {
		for _, f := range hookUserNickChange.regged {
			f(pkg, u, oldnick, newnick)
		}
	}
}

func runUserDataChangesHooks(pkg string, source, target *User, changes []DataChange, old []string) {
	for _, f := range hookUserDataChanges.all {
		f(pkg, source, target, changes, old)
	}

	if target.Registered() {
		for _, f := range hookUserDataChanges.regged {
			f(pkg, source, target, changes, old)
		}
	}
}

func runUserDeleteHooks(pkg string, source, target *User, message string, regged bool) {
	for _, f := range hookUserDelete.all {
		f(pkg, source, target, message)
	}

	if regged {
		for _, f := range hookUserDelete.regged {
			f(pkg, source, target, message)
		}
	}
}

func runUserDataChangeHooks(pkg string, source, target *User, name, oldvalue, newvalue string) {
	for _, f := range hookDataChange[name].all {
		f(pkg, source, target, oldvalue, newvalue)
	}

	if target.Registered() {
		for _, f := range hookDataChange[name].regged {
			f(pkg, source, target, oldvalue, newvalue)
		}
	}
}

func runUserMessageHooks(pkg string, source, target *User, message []byte, t string) {
	for _, f := range hookUserMessage[t] {
		f(pkg, source, target, message)
	}
}
