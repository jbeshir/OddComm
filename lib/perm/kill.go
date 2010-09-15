package perm

import "os"

import "oddcomm/src/core"


var checkKill *hook


func init() {
	// Add the core permissions for disconnecting a user.
	HookKill(noKilling)
}


// HookKill adds the given hook to CheckKill checks.
// The hook receives the source and target users.
// It should return a number indicating granted or denied permission, and the
// level of it. If the number is negative, err should be non-nil and indicate
// why. See package comment for permission levels.
func HookKill(h func(*core.User, *core.User) (int, os.Error)) {
	hookAdd(&checkKill, h)
}


// CheckKill tests whether the given user can disconnect the other given user.
func CheckKill(source, target *core.User) (bool, os.Error) {
	perm, err := CheckKillPerm(source, target)
	return perm > 0, err
}

// CheckKillPerm returns the full permissions value for CheckKill.
func CheckKillPerm(source, target *core.User) (int, os.Error) {
	f := func(f interface{}) (int, os.Error) {
		h, ok := f.(func(*core.User, *core.User) (int, os.Error))
		if ok && h != nil {
			return h(source, target)
		}
		return 0, nil
	}

	return checkKill.run(f, false)
}


// Permit no one without oper powers to kill another user.
func noKilling(source, target *core.User) (int, os.Error) {
	return -1000000, os.NewError("You are not a server operator.")
}
