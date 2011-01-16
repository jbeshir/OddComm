package perm

import "os"

import "oddcomm/src/core"


var checkUserMsg = make(map[string][]interface{})
var checkUserMsgAll[]interface{}
var checkChanMsg = make(map[string]map[string][]interface{})
var checkChanMsgAll = make(map[string][]interface{})


func init() {
	// Are the core permissions for sending messages to users.
	HookUserMsg(false, "invite", externalInvite)
	HookUserMsg(false, "invite", stupidInvite)

	// Add the core permissions for speaking in channels.
	HookChanMsg(true, "", "", externalMsg)
	HookChanMsg(true, "", "", muteBanned)
	HookChanMsg(true, "", "", moderated)
	HookChanMsg(true, "", "", voiceOverride)
	HookChanMsg(true, "", "", opMsgOverride)
}


// HookUserMsg adds the given hook to CheckUserMsg checks.
// The hook receives the source, the target, and the message.
// It should return a number indicating granted or denied permission, and the
// level of it. If the number is negative, err should be non-nil and indicate
// why. See package comment for permission levels.
// If all is true, the hook is called for all types of message. Otherwise, t is
// the type of message it wants to affect.
func HookUserMsg(all bool, t string, f func(string, *core.User, *core.User, []byte) (int, os.Error)) {
	if all {
		checkUserMsgAll = append(checkUserMsgAll, f)
	} else {
		checkUserMsg[t] = append(checkUserMsg[t], f)
	}
}

// HookChanMsg adds the given hook to CheckChanMsg checks.
// The hook receives the source, the target, and the message.
// It should return a number indicating granted or denied permission, and the
// level of it. If the number is negative, err should be non-nil and indicate
// why. See package comment for permission levels.
// If all is true, the hook is called for all types of message. Otherwise, t is
// the type of message it wants to affect.
func HookChanMsg(all bool, chantype, t string, f func(string, *core.User, *core.Channel, []byte) (int, os.Error)) {
	if all {
		checkChanMsgAll[chantype] = append(checkChanMsgAll[chantype], f)
	} else {
		if checkChanMsg[chantype] == nil {
			checkChanMsg[chantype] = make(map[string][]interface{})
		}
		checkChanMsg[chantype][t] = append(checkChanMsg[chantype][t], f)
	}
}

// CheckUserMsg tests whether the given user can PM the given target, with
// the given message and message type.
func CheckUserMsg(pkg string, source, target *core.User, message []byte, t string) (bool, os.Error) {
	perm, err := CheckUserMsgPerm(pkg, source, target, message, t)
	return perm > 0, err
}

// CheckUserMsgPerm returns the full permissions value for CheckUserMsg.
func CheckUserMsgPerm(pkg string, source, target *core.User, message []byte, t string) (int, os.Error) {
	f := func(h interface{}) (int, os.Error) {
		f, ok := h.(func(string, *core.User, *core.User, []byte) (int, os.Error))
		if ok && f != nil {
			return f(pkg, source, target, message)
		}
		return 0, nil
	}

	lists := make([][]interface{}, 2)
	lists[0] = checkUserMsgAll
	lists[1] = checkUserMsg[t]
	return runPermHooksSlice(lists, f, true)
}


// CheckChanMsg tests whether the given user can message the given channel,
// with the given message and message type.
func CheckChanMsg(source *core.User, target *core.Channel, message []byte, t string) (bool, os.Error) {
	perm, err := CheckChanMsgPerm(source, target, message, t)
	return perm > 0, err
}

// CheckChanMsgPerm returns the full permissions value for CheckChanMsg.
func CheckChanMsgPerm(source *core.User, target *core.Channel, message []byte, t string) (int, os.Error) {
	f := func(f interface{}) (int, os.Error) {
		h, ok := f.(func(*core.User, *core.Channel, []byte) (int, os.Error))
		if ok && h != nil {
			return h(source, target, message)
		}
		return 0, nil
	}

	chantype := target.Type()
	allList := checkChanMsgAll[chantype]
	var typeList []interface{}
	if checkChanMsg[chantype] != nil {
		typeList = checkChanMsg[chantype][t]
	}

	lists := make([][]interface{}, 2)
	lists[0] = allList
	lists[1] = typeList
	return runPermHooksSlice(lists, f, true)
}
