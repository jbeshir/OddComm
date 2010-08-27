package perm

import "os"

import "oddircd/src/core"


var checkNick *hook


// HookCheckNick adds the given hook to CheckNick checks.
// It should return a number indicating granted or denied permission, and the
// level of it. See package comment for permission levels.
func HookCheckNick(h func(*core.User, string) (int, os.Error)) {
	hookAdd(&checkNick, h)
}

// CheckNick tests whether the given user can change to the given nick.
// Note that this does not prevent collisions; these can only be checked for by
// checking SetNick's return value.
func CheckNick(u *core.User, nick string) (bool, os.Error) {
	perm, err := CheckNickPerm(u, nick)
	return perm > 0, err
}

// CheckNickPerm returns the full permissions value for CheckNick.
func CheckNickPerm(u *core.User, nick string) (int, os.Error) {
	return checkNick.run(func(f interface{}) (int, os.Error) {
		if h, ok := f.(func(*core.User, string) (int, os.Error)); ok && h != nil {
			return h(u, nick)
		}
		return 0, nil
	}, true)
}

