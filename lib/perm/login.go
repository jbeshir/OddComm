package perm

import "os"

import "oddcomm/src/core"


var checkLogin []interface{}


// HookCheckLogin adds the given hook to CheckLogin checks.
// The hook receives the user, an account name (may be "", for unspecified),
// the type of authentication information offered, and the authentication data
// as a string itself.
//
// It should return a number indicating granted or denied permission, and the
// level of it. 100 is standard to permit login, 10000 for specially blocking
// login (additional restrictions, too many attempts, other such).

// If the number is negative, err should be non-nil and indicate why. If it is
// positive, as a special case, err should be non-nil and convert to the
// account name to which login is granted.
func HookCheckLogin(f func(string, *core.User, string, string, string) (int, os.Error)) {
	checkLogin = append(checkLogin, f)
}


// CheckLogin tests whether the given user can login using the given
// username, authentication data type, and authentication data string.
func CheckLogin(pkg string, u *core.User, user, authtype, auth string) (bool, os.Error) {
	perm, err := CheckLoginPerm(pkg, u, user, authtype, auth)
	return perm > 0, err
}

// CheckLoginPerm returns the full permissions value for CheckLogin.
func CheckLoginPerm(pkg string, u *core.User, user, authtype, auth string) (int, os.Error) {
	f := func(h interface{}) (int, os.Error) {
		f, ok := h.(func(string, *core.User, string, string, string) (int, os.Error))
		if ok && f != nil {
			return f(pkg, u, user, authtype, auth)
		}
		return 0, nil
	}

	return runPermHooks(checkLogin, f, false)
}
