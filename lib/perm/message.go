package perm

import "os"

import "oddcomm/src/core"


var checkUserMsg map[string]**hook
var checkUserMsgAll *hook
var checkChanMsg map[string]map[string]**hook
var checkChanMsgAll map[string]**hook


func init() {
	checkUserMsg = make(map[string]**hook)
	checkChanMsg = make(map[string]map[string]**hook)
	checkChanMsgAll = make(map[string]**hook)

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

// HookChanMsg adds the given hook to CheckChanMsg checks.
// The hook receives the source, the target, and the message.
// It should return a number indicating granted or denied permission, and the
// level of it. If the number is negative, err should be non-nil and indicate
// why. See package comment for permission levels.
// If all is true, the hook is called for all types of message. Otherwise, t is
// the type of message it wants to affect.
func HookChanMsg(all bool, chantype, t string, h func(*core.User, *core.Channel, []byte) (int, os.Error)) {
	if all {
		if checkChanMsgAll[chantype] == nil {
			checkChanMsgAll[chantype] = new(*hook)
		}
		hookAdd(checkChanMsgAll[chantype], h)
	} else {
		if checkChanMsg[chantype] == nil {
			checkChanMsg[chantype] = make(map[string]**hook)
		}
		if checkChanMsg[chantype][t] == nil {
			checkChanMsg[chantype][t] = new(*hook)
		}
		hookAdd(checkChanMsg[chantype][t], h)
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
		h, ok := f.(func(*core.User, *core.User,
		                 []byte) (int, os.Error))
		if ok && h != nil {
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


// CheckChanMsg tests whether the given user can message the given channel,
// with the given message and message type.
func CheckChanMsg(source *core.User, target *core.Channel, message []byte,
                  t string) (bool, os.Error) {
	perm, err := CheckChanMsgPerm(source, target, message, t)
	return perm > 0, err
}

// CheckChanMsgPerm returns the full permissions value for CheckChanMsg.
func CheckChanMsgPerm(source *core.User, target *core.Channel, message []byte,
                      t string) (int, os.Error) {
	f := func(f interface{}) (int, os.Error) {
		h, ok := f.(func(*core.User, *core.Channel,
		                 []byte) (int, os.Error))
		if ok && h != nil {
			return h(source, target, message)
		}
		return 0, nil
	}

	chantype := target.Type()
	allList := checkChanMsgAll[chantype]
	var typeList **hook
	if checkChanMsg[chantype] != nil {
		typeList = checkChanMsg[chantype][t]
	}

	if allList != nil && typeList != nil {
		var lists [2]*hook
		lists[0] = *allList
		lists[1] = *typeList
		return runPermHookLists(lists[0:], f, true)
	}
	if allList != nil {
		return (*allList).run(f, true)
	}
	if typeList != nil {
		return (*typeList).run(f, true)
	}

	return 1, nil
}
