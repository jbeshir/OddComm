package perm

import "os"

import "oddircd/src/core"


var checkChanViewData map[string]**hook
var checkMemberViewData map[string]**hook

func init() {
	checkChanViewData = make(map[string]**hook)
	checkMemberViewData = make(map[string]**hook)
}


// HookCheckChanViewData adds the given hook to CheckChanViewData checks.
// The hook receives the user, the target channel, and the name of the data.
// It should return a number indicating granted or denied permission, and the
// level of it. If the number is negative, err should be non-nil and indicate
// why. See package comment for permission levels.
func HookCheckChanViewData(chantype string, h func(*core.User, *core.Channel, string) (int, os.Error)) {
	if checkChanViewData[chantype] == nil {
		checkChanViewData[chantype] = new(*hook)
	}
	hookAdd(checkChanViewData[chantype], h)
}

// HookCheckMemberViewData adds the given hook to CheckMemberViewData checks.
// The hook receives the user, the target channel, and the name of the data.
// It should return a number indicating granted or denied permission, and the
// level of it. If the number is negative, err should be non-nil and indicate
// why. See package comment for permission levels.
func HookCheckMemberViewData(chantype string, h func(*core.User, *core.Membership, string) (int, os.Error)) {
	if checkMemberViewData[chantype] == nil {
		checkMemberViewData[chantype] = new(*hook)
	}
	hookAdd(checkMemberViewData[chantype], h)
}


// CheckChanViewData tests whether the given user can view hidden data, such as
// the ban and unrestriction lists, on a channel.
func CheckChanViewData(u *core.User, ch *core.Channel, name string) (bool, os.Error) {
	perm, err := CheckChanViewDataPerm(u, ch, name)
	return perm > 0, err
}

// CheckChanViewDataPerm returns the full permissions value for
// CheckChanViewData.
func CheckChanViewDataPerm(u *core.User, ch *core.Channel, name string) (int, os.Error) {
	f := func(f interface{}) (int, os.Error) {
		h, ok := f.(func(*core.User, *core.Channel, string) (int, os.Error))
		if ok && h != nil {
			return h(u, ch, name)
		}
		return 0, nil
	}

	list := checkChanViewData[ch.Type()]
	if list != nil {
		return (*list).run(f, false)
	}
	return -1, os.NewError("You do not have permission to view channel hidden data.")
}

// CheckMemberViewData tests whether the given user can view flags of the given
// member of the channel.
func CheckMemberViewData(u *core.User, m *core.Membership, name string) (bool, os.Error) {
	perm, err := CheckMemberViewDataPerm(u, m, name)
	return perm > 0, err
}

// CheckMemberViewDataPerm returns the full permissions value for
// CheckMemberViewData.
func CheckMemberViewDataPerm(u *core.User, m *core.Membership, name string) (int, os.Error) {
	f := func(f interface{}) (int, os.Error) {
		h, ok := f.(func(*core.User, *core.Membership, string) (int, os.Error))
		if ok && h != nil {
			return h(u, m, name)
		}
		return 0, nil
	}

	list := checkMemberViewData[m.Channel().Type()]
	if list != nil {
		return (*list).run(f, false)
	}
	return -1, os.NewError("You do not have permission to view this member's flags.")
}


// CheckChanView tests whether the given user can see the given channel.
// This is not at present hookable.
func CheckChanView(u *core.User, ch *core.Channel) (bool, os.Error) {
	perm, err := CheckChanViewPerm(u, ch)
	return perm > 0, err
}

// CheckChanViewPerm returns the full permissions value for CheckChanView.
// This is not at present hookable.
func CheckChanViewPerm(u *core.User, ch *core.Channel) (int, os.Error) {

	if m := ch.GetMember(u); m != nil {
		return 100, nil
	}

	if ch.Data("hidden") == "" {
		return 1, nil
	}

	return -1, os.NewError("Channel is hidden.")
}
