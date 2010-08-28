package perm

import "os"
import "strings"
import "unicode"

import "oddcomm/src/core"


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

	// Block invalid utf8 from user specified strings.
	var dataNoInvalid = func(_, _ *core.User, _, s string) (int, os.Error) {
		if strings.IndexRune(s, unicode.ReplacementChar) != -1 {
			return -1e9, os.NewError("Invalid unicode specified.")
		}
		return 0, nil
	}
	HookCheckNick(func(_ *core.User, s string) (int, os.Error) {
		if strings.IndexRune(s, unicode.ReplacementChar) != -1 {
			return -1e9, os.NewError("Invalid unicode specified.")
		}
		return 0, nil
	})
	HookCheckUserData("ident", dataNoInvalid)
	HookCheckUserData("realname", dataNoInvalid)

	// Add core nick validation.
	// This only restricts the absolute minimum, as there is no way to
	// override this via another module.
	HookCheckNick(func(u *core.User, nick string) (int, os.Error) {
		// Block nicknames ambiguous with a unique ID, starting with a
		// number, and nine characters long, unless they are that
		// user's unique ID.
		if len(nick) == 9 && nick[0] < 58 && nick[0] > 47 {
			if nick != u.ID() {
				return -1e9, os.NewError("Nickname ambiguous with UIDs.")
			}
		}
		return 0, nil
	})
}
