package perm

import "oddircd/src/core"


var checkPM map[string]**hook
var checkPMAll *hook


func init() {
	checkPM = make(map[string]**hook)
}

// HookPermPM adds the given hook to PermPM checks.
// The hook receives the source, the target, and the message.
// It should return a number indicating granted or denied permission, and the
// level of it. See package comment for permission levels.
// If all is true, the hook is called for all types of message. Otherwise, t is
// the type of message it wants to affect.
func HookPermPM(all bool, t string,
                h func(*core.User, *core.User, []byte) int) {
	
	if all {
		hookAdd(&checkPMAll, h)
	} else {
		if checkPM[t] == nil {
			checkPM[t] = new(*hook)
		}
		hookAdd(checkPM[t], h)
	}
}

// CheckPM tests whether the given user can PM the given target, with the given
// message and message type.
func CheckPM(source, target *core.User, message []byte, t string) bool {
	return CheckPMPerm(source, target, message, t) > 0
}

// ChckPMPerm returns the full permissions value for CheckPM.
func CheckPMPerm(source, target *core.User, message []byte, t string) int {
	f := func(f interface{}) int {
		if h, ok := f.(func(*core.User, *core.User, []byte) int); ok &&
				h != nil {
			return h(source, target, message)
		}
		return 0
	}

	if checkPM[t] == nil {
		return checkPMAll.run(f, true)
	}

	var lists [2]*hook
	lists[0] = checkPMAll
	lists[1] = *checkPM[t]
	return runPermHookLists(lists[0:], f, true)
}
