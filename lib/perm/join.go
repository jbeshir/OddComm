package perm

import "os"

import "oddcomm/src/core"


var checkJoin = make(map[string][]interface{})
var checkRemove = make(map[string][]interface{})


func init() {
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
func HookJoin(chantype string, f func(string, *core.User, *core.Channel) (int, os.Error)) {
	checkJoin[chantype] = append(checkJoin[chantype], f)
}

// HookRemove adds the given hook to CheckRemove checks.
// The hook receives the source, the target, and the channel.
// It should return a number indicating granted or denied permission, and the
// level of it. If the number is negative, err should be non-nil and indicate
// why. See package comment for permission levels.
func HookRemove(chantype string, f func(string, *core.User, *core.User, *core.Channel) (int, os.Error)) {
	checkRemove[chantype] = append(checkRemove[chantype], f)
}


// CheckJoin tests whether the given user can join the given channel.
func CheckJoin(pkg string, source *core.User, target *core.Channel) (bool, os.Error) {
	perm, err := CheckJoinPerm(pkg, source, target)
	return perm > 0, err
}

// CheckJoinPerm returns the full permissions value for CheckJoin.
func CheckJoinPerm(pkg string, source *core.User, ch *core.Channel) (int, os.Error) {
	f := func(h interface{}) (int, os.Error) {
		f, ok := h.(func(string, *core.User, *core.Channel) (int, os.Error))
		if ok && f != nil {
			return f(pkg, source, ch)
		}
		return 0, nil
	}

	return runPermHooks(checkJoin[ch.Type()], f, true)
}

// CheckRemove tests whether the given user can remove the given target from
// the given channel.
func CheckRemove(pkg string, source, target *core.User, ch *core.Channel) (bool, os.Error) {
	perm, err := CheckRemovePerm(pkg, source, target, ch)
	return perm > 0, err
}

// CheckRemovePerm returns the full permissions value for CheckRemove.
func CheckRemovePerm(pkg string, source, target *core.User, ch *core.Channel) (int, os.Error) {
	f := func(h interface{}) (int, os.Error) {
		f, ok := h.(func(string, *core.User, *core.User, *core.Channel) (int, os.Error))
		if ok && f != nil {
			return f(pkg, source, target, ch)
		}
		return 0, nil
	}

	return runPermHooks(checkRemove[ch.Type()], f, false)
}
