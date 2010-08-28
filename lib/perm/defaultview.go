package perm

import "os"

import "oddcomm/src/core"


// Permit all users in a channel to view the ban list.
// This is really only permitted because IRC requires it, and there's no sense
// extending permissions there which aren't allowed elsewhere.
func viewChanBans(u *core.User, ch *core.Channel, name string) (int, os.Error) {
	if name == "ban" && ch.GetMember(u) != nil {
		return 100, nil
	}
	return 0, nil
}

// Permit chanops to view hidden channel data.
func viewChanOpOverride(u *core.User, ch *core.Channel, _ string) (int, os.Error) {
	if HasOpFlag(u, ch, "viewdata") {
		return 10000, nil
	}
	return 0, nil
}

// Permit chanops to view the flags of members.
func viewMemberOpOverride(u *core.User, m *core.Membership, _ string) (int, os.Error) {
	if HasOpFlag(u, m.Channel(), "viewmembers") {
		return 10000, nil
	}
	return 0, nil
}

// Permit a user to view their own flags.
func viewSelfOverride(u *core.User, m *core.Membership, _ string) (int, os.Error) {
	if u == m.User() {
		return 10000, nil
	}
	return 0, nil
}
