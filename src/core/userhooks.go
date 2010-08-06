package core

var initialRegcount int

var hookNewUser hooklist
var hookRegistered hooklist
var hookNickChange hooklist
var hookRemoved hooklist

var dataChangeHooks map[string]*hooklist


func init() {
	dataChangeHooks = make(map[string]*hooklist)

	IncrementRegcount()
	HookNickChange(func(u *User, oldnick string) {
		if oldnick == "" {
			u.DecrementRegcount()
		}
	}, true)
}

func IncrementRegcount() {
	initialRegcount++
}

func HookNewUser(f func(*User)) {
	hookNewUser.add(f, true)
}

func HookRegistered(f func(*User)) {
	hookRegistered.add(f, false)
}

func HookNickChange(f func(*User, string), unregged bool) {
	hookNickChange.add(f, unregged)
}

func HookRemoved(f func(*User, string), unregged bool) {
	hookRemoved.add(f, unregged)
}

func HookDataChange(data string, f func(*User, string), unregged bool) {
	if dataChangeHooks[data] == nil {
		dataChangeHooks[data] = new(hooklist)
	}

	dataChangeHooks[data].add(f, unregged)		
}


func runNewUserHooks(u *User) {
	hookNewUser.run(func(f interface{}) {
		if h := f.(func(*User)); h != nil {
			h(u)
		}
	}, false)
}

func runRegisteredHooks(u *User) {
	hookRegistered.run(func(f interface{}) {
		if h, ok := f.(func(*User)); ok && h != nil {
			h(u)
		}
	}, true)
}

func runNickChangeHooks(u *User, oldnick string) {
	hookNickChange.run(func(f interface{}) {
		if h, ok := f.(func(*User, string)); ok && h != nil {
			h(u, oldnick)
		}
	}, u.Registered())
}

func runRemovedHooks(u *User, message string) {
	hookRemoved.run(func(f interface{}) {
		if h, ok := f.(func(*User, string)); ok && h != nil {
			h(u, message)
		}
	}, u.Registered())
}

func runDataChangeHooks(u *User, name string, oldvalue string) {
	if dataChangeHooks[name] == nil { return }
	dataChangeHooks[name].run(func(f interface{}) {
		if h, ok := f.(func(*User, string)); ok && h != nil {
			h(u, oldvalue)
		}
	}, u.Registered())
}
