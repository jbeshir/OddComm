package perm

import "os"

import "oddcomm/src/core"


var checkKill []interface{}


func init() {
	// Add the core permissions for disconnecting a user.
	HookKill(noKilling)
}


// HookKill adds the given hook to CheckKill checks.
// The hook receives the source and target users.
// It should return a number indicating granted or denied permission, and the
// level of it. If the number is negative, err should be non-nil and indicate
// why. See package comment for permission levels.
func HookKill(f func(string, *core.User, *core.User) (int, os.Error)) {
	checkKill = append(checkKill, f)
}


// CheckKill tests whether the given user can disconnect the other given user.
func CheckKill(pkg string, source, target *core.User) (bool, os.Error) {
	perm, err := CheckKillPerm(pkg, source, target)
	return perm > 0, err
}

// CheckKillPerm returns the full permissions value for CheckKill.
func CheckKillPerm(pkg string, source, target *core.User) (int, os.Error) {
	f := func(h interface{}) (int, os.Error) {
		f, ok := h.(func(string, *core.User, *core.User) (int, os.Error))
		if ok && f != nil {
			return f(pkg, source, target)
		}
		return 0, nil
	}

	return runPermHooks(checkKill, f, false)
}


// Permit no one without oper powers to kill another user.
func noKilling(_ string, source, target *core.User) (int, os.Error) {
	return -1000000, os.NewError("You are not a server operator.")
}
