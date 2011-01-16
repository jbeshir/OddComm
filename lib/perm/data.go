package perm

import "os"
import "strings"

import "oddcomm/src/core"


var checkUserData = make(map[string][]interface{})

var checkChanData = make(map[string]map[string][]interface{})

var checkMemberData = make(map[string]map[string][]interface{})


// HookCheckUserData adds the given hook to CheckUserData checks for data with
// either the exact same name, or beginning with that prefix followed by a
// space. The hook receives the user, the target user, and the name and value
// of the data being set.
// It should return a number indicating granted or denied permission, and the
// level of it. If the number is negative, err should be non-nil and indicate
// why. See package comment for permission levels.
func HookCheckUserData(name string, f func(string, *core.User, *core.User, string, string) (int, os.Error)) {
	checkUserData[name] = append(checkUserData[name], f)
}

// HookCheckChanData adds the given hook to CheckChanData checks for data with
// either the exact same name, or beginning with that prefix followed by a
// space. The hook receives the user, the channel, and the name and value of
// the data being set.
// It should return a number indicating granted or denied permission, and the
// level of it. If the number is negative, err should be non-nil and indicate
// why. See package comment for permission levels.
func HookCheckChanData(chantype, name string, f func(string, *core.User, *core.Channel, string, string) (int, os.Error)) {
	if checkChanData[chantype] == nil {
		checkChanData[chantype] = make(map[string][]interface{})
	}
	checkChanData[chantype][name] = append(checkChanData[chantype][name], f)
}

// HookCheckMemberData adds the given hook to CheckMemberData checks for data
// with either the exact same name, or beginning with that prefix followed by a
// space. The hook receives the user, the target membership, and the name and
// value of the data being set.
// It should return a number indicating granted or denied permission, and the
// level of it. If the number is negative, err should be non-nil and indicate
// why. See package comment for permission levels.
func HookCheckMemberData(chantype, name string, f func(string, *core.User, *core.Membership, string, string) (int, os.Error)) {
	if checkMemberData[chantype] == nil {
		checkMemberData[chantype] = make(map[string][]interface{})
	}
	checkMemberData[chantype][name] = append(checkMemberData[chantype][name], f)
}


// CheckUserData tests whether the given user can set data with the given name
// and value on the given user.
func CheckUserData(pkg string, source, target *core.User, name, value string) (bool, os.Error) {
	perm, err := CheckUserDataPerm(pkg, source, target, name, value)
	return perm > 0, err
}

// CheckUserDataPerm returns the full permissions value for CheckUserData.
func CheckUserDataPerm(pkg string, source, target *core.User, name, value string) (int, os.Error) {
	f := func(h interface{}) (int, os.Error) {
		f, ok := h.(func(string, *core.User, *core.User, string, string) (int, os.Error))
		if ok && f != nil {
			return f(pkg, source, target, name, value)
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
			prefix = prefix[:v]
		} else {
			break
		}
	}
	lists := make([][]interface{}, 0, len)

	prefix = name
	for {
		if v := checkUserData[prefix]; v != nil {
			lists = append(lists, v)
		}

		if v := strings.LastIndex(prefix, " "); v != -1 {
			prefix = prefix[:v]
		} else {
			break
		}
	}

	return runPermHooksSlice(lists, f, false)
}

// CheckChanData tests whether the given user can set data with the given name
// and value on the given channel.
func CheckChanData(pkg string, u *core.User, ch *core.Channel, name, value string) (bool, os.Error) {
	perm, err := CheckChanDataPerm(pkg, u, ch, name, value)
	return perm > 0, err
}

// CheckChanDataPerm returns the full permissions value for CheckChanData.
func CheckChanDataPerm(pkg string, u *core.User, ch *core.Channel, name, value string) (int, os.Error) {
	f := func(h interface{}) (int, os.Error) {
		f, ok := h.(func(string, *core.User, *core.Channel, string, string) (int, os.Error))
		if ok && f != nil {
			return f(pkg, u, ch, name, value)
		}
		return 0, nil
	}
	chantype := ch.Type()

	var prefix = name
	var len int
	for {
		if v := checkChanData[chantype][prefix]; v != nil {
			len++
		}

		if v := strings.LastIndex(prefix, " "); v != -1 {
			prefix = prefix[:v]
		} else {
			break
		}
	}
	lists := make([][]interface{}, 0, len)

	prefix = name
	for {
		if v := checkChanData[chantype][prefix]; v != nil {
			lists = append(lists, v)
		}

		if v := strings.LastIndex(prefix, " "); v != -1 {
			prefix = prefix[:v]
		} else {
			break
		}
	}

	return runPermHooksSlice(lists, f, false)
}

// CheckMemberData tests whether the given user can set data with the given
// name and value on the given membership entry.
func CheckMemberData(pkg string, u *core.User, m *core.Membership, name, value string) (bool, os.Error) {
	perm, err := CheckMemberDataPerm(pkg, u, m, name, value)
	return perm > 0, err
}

// CheckMemberDataPerm returns the full permissions value for CheckMemberData.
func CheckMemberDataPerm(pkg string, u *core.User, m *core.Membership, name, value string) (int, os.Error) {
	f := func(h interface{}) (int, os.Error) {
		f, ok := h.(func(string, *core.User, *core.Membership, string, string) (int, os.Error))
		if ok && h != nil {
			return f(pkg, u, m, name, value)
		}
		return 0, nil
	}
	chantype := m.Channel().Type()

	var prefix = name
	var len int
	for {
		if v := checkMemberData[chantype][prefix]; v != nil {
			len++
		}

		if v := strings.LastIndex(prefix, " "); v != -1 {
			prefix = prefix[:v]
		} else {
			break
		}
	}
	lists := make([][]interface{}, len)

	prefix = name
	for {
		if v := checkMemberData[chantype][prefix]; v != nil {
			lists = append(lists, v)
		}

		if v := strings.LastIndex(prefix, " "); v != -1 {
			prefix = prefix[:v]
		} else {
			break
		}
	}

	return runPermHooksSlice(lists, f, false)
}


// Permits ops with the given flag to set metadata with the given name, or
// beginning with the given name, on the given type of channel.
func PermitChanDataOp(chantype, flag, name string) {
	HookCheckChanData(chantype, name, func(_ string, u *core.User, ch *core.Channel, _, _ string) (int, os.Error) {
		if HasOpFlag(u, ch, flag) {
			return 10000, nil
		}
		return 0, nil
	})
}

// Permits ops with the given flag to set member metadata with the given name,
// or beginning with the given name, on the given type of channel.
func PermitMemberDataOp(chantype, flag, name string) {
	HookCheckMemberData(chantype, name, func(_ string, u *core.User, m *core.Membership, _, _ string) (int, os.Error) {
		if HasOpFlag(u, m.Channel(), flag) {
			return 10000, nil
		}
		return 0, nil
	})
}
