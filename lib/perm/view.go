package perm

import "os"

import "oddcomm/src/core"


var checkChanViewData = make(map[string][]interface{})
var checkMemberViewData = make(map[string][]interface{})

func init() {
	// Add core view restrictions/permissions.
	HookCheckChanViewData("", viewChanBans)
	HookCheckChanViewData("", viewChanOpOverride)
	HookCheckMemberViewData("", viewMemberOpOverride)
	HookCheckMemberViewData("", viewSelfOverride)
}


// HookCheckChanViewData adds the given hook to CheckChanViewData checks.
// The hook receives the user, the target channel, and the name of the data.
// It should return a number indicating granted or denied permission, and the
// level of it. If the number is negative, err should be non-nil and indicate
// why. See package comment for permission levels.
func HookCheckChanViewData(chantype string, f func(string, *core.User, *core.Channel, string) (int, os.Error)) {
	checkChanViewData[chantype] = append(checkChanViewData[chantype], f)
}

// HookCheckMemberViewData adds the given hook to CheckMemberViewData checks.
// The hook receives the user, the target channel, and the name of the data.
// It should return a number indicating granted or denied permission, and the
// level of it. If the number is negative, err should be non-nil and indicate
// why. See package comment for permission levels.
func HookCheckMemberViewData(chantype string, f func(string, *core.User, *core.Membership, string) (int, os.Error)) {
	checkMemberViewData[chantype] = append(checkMemberViewData[chantype], f)
}


// CheckChanViewData tests whether the given user can view hidden data, such as
// the ban and unrestriction lists, on a channel.
func CheckChanViewData(pkg string, u *core.User, ch *core.Channel, name string) (bool, os.Error) {
	perm, err := CheckChanViewDataPerm(pkg, u, ch, name)
	return perm > 0, err
}

// CheckChanViewDataPerm returns the full permissions value for
// CheckChanViewData.
func CheckChanViewDataPerm(pkg string, u *core.User, ch *core.Channel, name string) (int, os.Error) {
	f := func(h interface{}) (int, os.Error) {
		f, ok := h.(func(string, *core.User, *core.Channel, string) (int, os.Error))
		if ok && f != nil {
			return f(pkg, u, ch, name)
		}
		return 0, nil
	}

	return runPermHooks(checkChanViewData[ch.Type()], f, false)
}

// CheckMemberViewData tests whether the given user can view flags of the given
// member of the channel.
func CheckMemberViewData(pkg string, u *core.User, m *core.Membership, name string) (bool, os.Error) {
	perm, err := CheckMemberViewDataPerm(pkg, u, m, name)
	return perm > 0, err
}

// CheckMemberViewDataPerm returns the full permissions value for
// CheckMemberViewData.
func CheckMemberViewDataPerm(pkg string, u *core.User, m *core.Membership, name string) (int, os.Error) {
	f := func(h interface{}) (int, os.Error) {
		f, ok := h.(func(string, *core.User, *core.Membership, string) (int, os.Error))
		if ok && f != nil {
			return f(pkg, u, m, name)
		}
		return 0, nil
	}

	return runPermHooks(checkMemberViewData[m.Channel().Type()], f, false)
}


// CheckChanView tests whether the given user can see the given channel.
// This is not at present hookable.
func CheckChanView(pkg string, u *core.User, ch *core.Channel) (bool, os.Error) {
	perm, err := CheckChanViewPerm(pkg, u, ch)
	return perm > 0, err
}

// CheckChanViewPerm returns the full permissions value for CheckChanView.
// This is not at present hookable.
func CheckChanViewPerm(pkg string, u *core.User, ch *core.Channel) (int, os.Error) {

	if m := ch.GetMember(u); m != nil {
		return 100, nil
	}

	if ch.Data("hidden") == "" {
		return 1, nil
	}

	return -1, os.NewError("Channel is hidden.")
}
