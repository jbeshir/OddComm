package perm

import "os"

import "oddircd/src/core"


var checkUserMsg map[string]**hook
var checkUserMsgAll *hook
var checkChanMsg map[string]map[string]**hook
var checkChanMsgAll map[string]**hook


func init() {
	checkUserMsg = make(map[string]**hook)
	checkChanMsg = make(map[string]map[string]**hook)
	checkChanMsgAll = make(map[string]**hook)
}


// HookUserMsg adds the given hook to CheckUserMsg checks.
// The hook receives the source, the target, and the message.
// It should return a number indicating granted or denied permission, and the
// level of it. If the number is negative, err should be non-nil and indicate
// why. See package comment for permission levels.
// If all is true, the hook is called for all types of message. Otherwise, t is
// the type of message it wants to affect.
func HookUserMsg(all bool, t string,
                     h func(*core.User, *core.User, []byte) (int, os.Error)) {
	if all {
		hookAdd(&checkUserMsgAll, h)
	} else {
		if checkUserMsg[t] == nil {
			checkUserMsg[t] = new(*hook)
		}
		hookAdd(checkUserMsg[t], h)
	}
}

// CheckUserMsg tests whether the given user can PM the given target, with
// the given message and message type.
func CheckUserMsg(source, target *core.User, message []byte,
                      t string) (bool, os.Error) {
	perm, err := CheckUserMsgPerm(source, target, message, t)
	return perm > 0, err
}

// CheckUserMsgPerm returns the full permissions value for CheckUserMsg.
func CheckUserMsgPerm(source, target *core.User, message []byte,
                          t string) (int, os.Error) {
	f := func(f interface{}) (int, os.Error) {
		if h, ok := f.(func(*core.User, *core.User, []byte) (int, os.Error)); ok &&
				h != nil {
			return h(source, target, message)
		}
		return 0, nil
	}

	if checkUserMsg[t] == nil {
		return checkUserMsgAll.run(f, true)
	}

	var lists [2]*hook
	lists[0] = checkUserMsgAll
	lists[1] = *checkUserMsg[t]
	return runPermHookLists(lists[0:], f, true)
}
