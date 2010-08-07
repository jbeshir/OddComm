package client

import "strings"

import "oddircd/src/core"
import "oddircd/src/perm"

func init() {
	// Impose the IRC client module's limitations on nicks.
	// These are strictly what renders the protocol ambiguous, or will
	// very likely break clients.
	perm.HookValidateNick(func(u *core.User, nick string) int {

		// Do not permit space or comma anywhere in a nick.
		if strings.IndexAny(nick, " ,") != -1 {
			return -1e9
		}

		// Do not permit ! in a nick anywhere but the start.
		if strings.IndexRune(nick[1:], '!') != -1 {
			return -1e9
		}

		if nick[0] <= 127 {
			// Block nicks starting with one of the standard (or
			// particularly common) operator symbols.
			// While these shouldn't be any more bad than any
			// arbitrary unusual characters, clients do insist on
			// hardcoding prefixes, even non-standard ones.
			if strings.IndexRune("~&@%+", int(nick[0])) != -1 {
				return -1e9
			}
		}

		return 0
	})
}
