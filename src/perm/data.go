package perm

import "os"
import "strings"

import "oddircd/src/core"


var checkUserData map[string]**hook
var checkChanData map[string]map[string]**hook
var checkMemberData map[string]map[string]**hook


func init() {
	checkUserData = make(map[string]**hook)
	checkChanData = make(map[string]map[string]**hook)
	checkMemberData = make(map[string]map[string]**hook)
}


// HookCheckUserData adds the given hook to CheckUserData checks for data with
// either the exact same name, or beginning with that prefix followed by a
// space. The hook receives the user, the target user, and the name and value
// of the data being set.
// It should return a number indicating granted or denied permission, and the
// level of it. If the number is negative, err should be non-nil and indicate
// why. See package comment for permission levels.
func HookCheckUserData(name string, h func(*core.User, *core.User, string, string) (int, os.Error)) {
	if checkUserData[name] == nil {
		checkUserData[name] = new(*hook)
	}
	hookAdd(checkUserData[name], h)
}

// HookCheckChanData adds the given hook to CheckChanData checks for data with
// either the exact same name, or beginning with that prefix followed by a
// space. The hook receives the user, the channel, and the name and value of
// the data being set.
// It should return a number indicating granted or denied permission, and the
// level of it. If the number is negative, err should be non-nil and indicate
// why. See package comment for permission levels.
func HookCheckChanData(chantype, name string, h func(*core.User, *core.Channel, string, string) (int, os.Error)) {
	if checkChanData[chantype] == nil {
		checkChanData[chantype] = make(map[string]**hook)
	}
	if checkChanData[chantype][name] == nil {
		checkChanData[chantype][name] = new(*hook)
	}
	hookAdd(checkChanData[chantype][name], h)
}

// HookCheckMemberData adds the given hook to CheckMemberData checks for data
// with either the exact same name, or beginning with that prefix followed by a
// space. The hook receives the user, the target membership, and the name and
// value of the data being set.
// It should return a number indicating granted or denied permission, and the
// level of it. If the number is negative, err should be non-nil and indicate
// why. See package comment for permission levels.
func HookCheckMemberData(chantype, name string, h func(*core.User, *core.Membership, string, string) (int, os.Error)) {
	if checkMemberData[chantype] == nil {
		checkMemberData[chantype] = make(map[string]**hook)
	}
	if checkMemberData[chantype][name] == nil {
		checkMemberData[chantype][name] = new(*hook)
	}
	hookAdd(checkMemberData[chantype][name], h)
}


// CheckUserData tests whether the given user can set data with the given name
// and value on the given user.
func CheckUserData(source, target *core.User, name, value string) (bool, os.Error) {
	perm, err := CheckUserDataPerm(source, target, name, value)
	return perm > 0, err
}

// CheckUserDataPerm returns the full permissions value for CheckUserData.
func CheckUserDataPerm(source, target *core.User, name, value string) (int, os.Error) {
	f := func(f interface{}) (int, os.Error) {
		h, ok := f.(func(*core.User, *core.User, string,
		                 string) (int, os.Error))
		if ok && h != nil {
			return h(source, target, name, value)
		}
		return 0, nil
	}

	var prefix = name
	var len int
	for {
		if v := checkUserData[prefix]; v != nil {
			len++
		}

		if v := strings.LastIndex(prefix, " "); v != -1 {
			prefix = prefix[0:v]
		} else {
			break
		}
	}

	lists := make([]*hook, len)
	for i, prefix := 0, name; i < len; {
		if v := checkUserData[prefix]; v != nil {
			lists[i] = *v
			i++
		}

		prefix = prefix[0:strings.LastIndex(prefix, " ")]
	}

	return runPermHookLists(lists, f, false)
}

// CheckChanData tests whether the given user can set data with the given name
// and value on the given channel.
func CheckChanData(u *core.User, ch *core.Channel, name, value string) (bool, os.Error) {
	perm, err := CheckChanDataPerm(u, ch, name, value)
	return perm > 0, err
}

// CheckChanDataPerm returns the full permissions value for CheckChanData.
func CheckChanDataPerm(u *core.User, ch *core.Channel, name, value string) (int, os.Error) {
	f := func(f interface{}) (int, os.Error) {
		h, ok := f.(func(*core.User, *core.Channel, string,
		                 string) (int, os.Error))
		if ok && h != nil {
			return h(u, ch, name, value)
		}
		return 0, nil
	}

	if checkChanData[ch.Type()] == nil {
		return -1, os.NewError("Permission denied.")
	}

	var prefix = name
	var len int
	for {
		if v := checkChanData[ch.Type()][prefix]; v != nil {
			len++
		}

		if v := strings.LastIndex(prefix, " "); v != -1 {
			prefix = prefix[0:v]
		} else {
			break
		}
	}

	lists := make([]*hook, len)
	for i, prefix := 0, name; i < len; {
		if v := checkChanData[ch.Type()][prefix]; v != nil {
			lists[i] = *v
			i++
		}

		prefix = prefix[0:strings.LastIndex(prefix, " ")]
	}

	return runPermHookLists(lists, f, false)
}

// CheckMemberData tests whether the given user can set data with the given
// name and value on the given membership entry.
func CheckMemberData(u *core.User, m *core.Membership, name, value string) (bool, os.Error) {
	perm, err := CheckMemberDataPerm(u, m, name, value)
	return perm > 0, err
}

// CheckMemberDataPerm returns the full permissions value for CheckMemberData.
func CheckMemberDataPerm(u *core.User, m *core.Membership, name, value string) (int, os.Error) {
	f := func(f interface{}) (int, os.Error) {
		h, ok := f.(func(*core.User, *core.Membership, string,
		                 string) (int, os.Error))
		if ok && h != nil {
			return h(u, m, name, value)
		}
		return 0, nil
	}
	chantype := m.Channel().Type()

	if checkMemberData[chantype] == nil {
		return -1, os.NewError("Permission denied.")
	}

	var prefix = name
	var len int
	for {
		if v := checkMemberData[chantype][prefix]; v != nil {
			len++
		}

		if v := strings.LastIndex(prefix, " "); v != -1 {
			prefix = prefix[0:v]
		} else {
			break
		}
	}

	lists := make([]*hook, len)
	for i, prefix := 0, name; i < len; {
		if v := checkMemberData[chantype][prefix]; v != nil {
			lists[i] = *v
			i++
		}

		prefix = prefix[0:strings.LastIndex(prefix, " ")]
	}

	return runPermHookLists(lists, f, false)
}
