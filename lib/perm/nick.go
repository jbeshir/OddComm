package perm

import "os"

import "oddcomm/src/core"


var checkNick []interface{}


// HookCheckNick adds the given hook to CheckNick checks.
// It should return a number indicating granted or denied permission, and the
// level of it. See package comment for permission levels.
func HookCheckNick(f func(string, *core.User, string) (int, os.Error)) {
	checkNick = append(checkNick, f)
}

// CheckNick tests whether the given user can change to the given nick.
// Note that this does not prevent collisions; these can only be checked for by
// checking SetNick's return value.
func CheckNick(pkg string, u *core.User, nick string) (bool, os.Error) {
	perm, err := CheckNickPerm(pkg, u, nick)
	return perm > 0, err
}

// CheckNickPerm returns the full permissions value for CheckNick.
func CheckNickPerm(pkg string, u *core.User, nick string) (int, os.Error) {
	return runPermHooks(checkNick, func(h interface{}) (int, os.Error) {
		if f, ok := h.(func(string, *core.User, string) (int, os.Error)); ok && h != nil {
			return f(pkg, u, nick)
		}
		return 0, nil
	},true)
}
