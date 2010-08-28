package perm

import "os"

import "oddcomm/src/core"


var checkJoin map[string]**hook
var checkRemove map[string]**hook


func init() {
	checkJoin = make(map[string]**hook)
	checkRemove = make(map[string]**hook)
	
	// Add the core permissions for joining channels.
	HookJoin("", joinBanned) 
	HookJoin("", inviteOnly)

	// Add the core permissions for removing users from channels.
	HookRemove("", selfOverride)
	HookRemove("", opKickImmune)
	HookRemove("", opKickOverride)
}


// HookJoin adds the given hook to CheckJoin checks.
// The hook receives the user and the target channel.
// It should return a number indicating granted or denied permission, and the
// level of it. If the number is negative, err should be non-nil and indicate
// why. See package comment for permission levels.
func HookJoin(chantype string, h func(*core.User, *core.Channel) (int, os.Error)) {
	if checkJoin[chantype] == nil {
		checkJoin[chantype] = new(*hook)
	}
	hookAdd(checkJoin[chantype], h)
}

// HookRemove adds the given hook to CheckRemove checks.
// The hook receives the source, the target, and the channel.
// It should return a number indicating granted or denied permission, and the
// level of it. If the number is negative, err should be non-nil and indicate
// why. See package comment for permission levels.
func HookRemove(chantype string, h func(*core.User, *core.User, *core.Channel) (int, os.Error)) {
	if checkRemove[chantype] == nil {
		checkRemove[chantype] = new(*hook)
	}
	hookAdd(checkRemove[chantype], h)
}


// CheckJoin tests whether the given user can join the given channel.
func CheckJoin(source *core.User, target *core.Channel) (bool, os.Error) {
	perm, err := CheckJoinPerm(source, target)
	return perm > 0, err
}

// CheckJoinPerm returns the full permissions value for CheckJoin.
func CheckJoinPerm(source *core.User, ch *core.Channel) (int, os.Error) {
	f := func(f interface{}) (int, os.Error) {
		h, ok := f.(func(*core.User, *core.Channel) (int, os.Error))
		if ok && h != nil {
			return h(source, ch)
		}
		return 0, nil
	}

	list := checkJoin[ch.Type()]
	if list != nil {
		return (*list).run(f, true)
	}
	return 1, nil
}

// CheckRemove tests whether the given user can remove the given target from
// the given channel.
func CheckRemove(source, target *core.User, ch *core.Channel) (bool, os.Error) {
	perm, err := CheckRemovePerm(source, target, ch)
	return perm > 0, err
}

// CheckRemovePerm returns the full permissions value for CheckRemove.
func CheckRemovePerm(source, target *core.User, ch *core.Channel) (int, os.Error) {
	f := func(f interface{}) (int, os.Error) {
		h, ok := f.(func(*core.User, *core.User, *core.Channel) (int, os.Error))
		if ok && h != nil {
			return h(source, target, ch)
		}
		return 0, nil
	}

	list := checkRemove[ch.Type()]
	if list != nil {
		return (*list).run(f, false)
	}
	return 1, nil
}
