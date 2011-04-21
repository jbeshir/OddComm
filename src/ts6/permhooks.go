package ts6

import "os"

import "oddcomm/src/core"
import "oddcomm/lib/perm"

func init() {
	// Block nick changes of our clients not originating from their server.
	perm.HookCheckNick(func(pkg string, u *core.User, nick string) (int, os.Error) {
		if u.Owner() == me {
			return -1e9, os.NewError("Cannot change remote user nicks.")
		}
		return 0, nil
	})
}
