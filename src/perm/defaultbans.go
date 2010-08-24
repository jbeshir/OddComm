package perm

import "oddircd/src/core"


func init() {
	// Add the default ban types.
	AddBanType("account", func (u *core.User, mask string) bool {
		if acc := u.Data("account"); acc != "" {
			return GMatch(acc, mask)
		}
		return false
	})
	AddBanType("host", func (u *core.User, mask string) bool {
		nuh := u.Nick() + "!" + u.GetIdent() + "@" + u.GetHostname()
		return GMatch(nuh, mask)
	})
}
