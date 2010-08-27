package perm

import "os"

import "oddircd/src/core"


func init() {
	// Permit chanops with the restrict flag to set (un)restrict metadata.
	// Also, to hide/unhide the channel.
	PermitChanDataOp("", "restrict", "restrict")
	PermitChanDataOp("", "restrict", "unrestrict")
	PermitChanDataOp("", "restrict", "hidden")

	// Permit chanops with the ban flag to set bans and ban exceptions.
	PermitChanDataOp("", "ban", "ban")
	PermitChanDataOp("", "ban", "banexception")

	// Permit chanops with the mark flag to voice users.
	PermitMemberDataOp("", "mark", "voiced")

	// Permit chanops with the op flag to op/deop users.
	// This is not restricted to flags this user has themselves.
	PermitMemberDataOp("", "op", "op")

	// Permit chanops to deop themselves.
	HookCheckMemberData("", "op", func(u *core.User, m *core.Membership, name, value string) (int, os.Error) {
		if u == m.User() && value == "" {
			return 10000, nil
		}
		return 0, nil
	})
}
