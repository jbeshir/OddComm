package perm

import "os"
import "strings"
import "unicode"

import "oddircd/src/core"


var validNick *hook
var validIdent *hook
var validRealname *hook


// HookValidateNick adds the given hook to ValidateNick checks.
// It should return a number indicating granted or denied permission, and the
// level of it. See package comment for permission levels.
func HookValidateNick(h func(*core.User, string) (int, os.Error)) {
	hookAdd(&validNick, h)
}

// HookValidateIdent adds the given hook to ValidateIdent checks.
// It should return a number indicating granted or denied permission, and the
// level of it. See package comment for permission levels.
func HookValidateIdent(h func(*core.User, string) (int, os.Error)) {
	hookAdd(&validIdent, h)
}

// HookValidateRealname adds the given hook to ValidateRealname checks.
// It should return a number indicating granted or denied permission, and the
// level of it. See package comment for permission levels.
func HookValidateRealname(h func(*core.User, string) (int, os.Error)) {
	hookAdd(&validRealname, h)
}

// ValidateNick tests whether the given user can use the given nick.
// Note that this does not prevent collisions; these can only be checked for by
// checking SetNick's return value.
func ValidateNick(u *core.User, nick string) (bool, os.Error) {
	perm, err := ValidateNickPerm(u, nick)
	return perm > 0, err
}

// ValidateNickPerm returns the full permissions value for ValidateNick.
func ValidateNickPerm(u *core.User, nick string) (int, os.Error) {
	return validNick.run(func(f interface{}) (int, os.Error) {
		if h, ok := f.(func(*core.User, string) (int, os.Error)); ok && h != nil {
			return h(u, nick)
		}
		return 0, nil
	}, true)
}

// ValidateIdent tests whether the given user can use the given ident.
func ValidateIdent(u *core.User, ident string) (bool, os.Error) {
	perm, err := ValidateIdentPerm(u, ident)
	return perm > 0, err
}

// ValidateIdentPerm returns the full permissions value for ValidateIdent.
func ValidateIdentPerm(u *core.User, ident string) (int, os.Error) {
	return validIdent.run(func(f interface{}) (int, os.Error) {
		if h, ok := f.(func(*core.User, string) (int, os.Error)); ok && h != nil {
			return h(u, ident)
		}
		return 0, nil
	}, true)
}

// ValidateRealname tests whether the given user can use the given name.
func ValidateRealname(u *core.User, name string) (bool, os.Error) {
	perm, err := ValidateRealnamePerm(u, name)
	return perm > 0, err
}

// ValidateRealnamePerm returns the full permissions value for ValidateRealname.
func ValidateRealnamePerm(u *core.User, name string) (int, os.Error) {
	return validRealname.run(func(f interface{}) (int, os.Error) {
		if h, ok := f.(func(*core.User, string) (int, os.Error)); ok && h != nil {
			return h(u, name)
		}
		return 0, nil
	}, true)
}


func init() {
	// Block invalid utf8 from everything.
	// We don't like binary gibberish.
	var noInvalid = func(u *core.User, s string) (int, os.Error) {
		if strings.IndexRune(s, unicode.ReplacementChar) != -1 {
			return -1e9, os.NewError("Invalid unicode specified.")
		}
		return 0, nil
	}
	HookValidateNick(noInvalid)
	HookValidateIdent(noInvalid)
	HookValidateRealname(noInvalid)

	// Add core nick validation.
	// This only restricts the absolute minimum, as there is no way to
	// override this via another module.
	HookValidateNick(func(u *core.User, nick string) (int, os.Error) {
		// Block nicknames ambiguous with a unique ID, starting with a
		// number, and nine characters long, unless they are that
		// user's unique ID.
		if len(nick) == 9 && nick[0] < 58 && nick[0] > 47 {
			if nick != u.ID() {
				return -1e9, os.NewError("Nickname ambiguous with UIDs.")
			}
		}
		return 0, nil
	})
}
