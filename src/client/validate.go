package client

import "os"
import "strings"

import "oddcomm/src/core"
import "oddcomm/lib/perm"

func init() {
	// Impose the IRC client module's limitations on nicks.
	// These are strictly what renders the protocol ambiguous, or will
	// very likely break clients.
	perm.HookCheckNick(func(_ string, u *core.User, nick string) (int, os.Error) {

		// Do not permit space or comma anywhere in a nick.
		if strings.IndexAny(nick, " ,") != -1 {
			return -1e9, os.NewError("Nickname contains space or comma.")
		}

		// Do not permit ! anywhere in a nick.
		// The start could hypothetically be legal, but breaks clients.
		if strings.IndexRune(nick, '!') != -1 {
			return -1e9, os.NewError("Nickname contains ! character.")
		}

		if nick[0] <= 127 {
			// Block nicks starting with the channel prefix.
			if strings.IndexRune("#", int(nick[0])) != -1 {
				return -1e9, os.NewError("Nickname starts with illegal character.")
			}

			// Block nicks starting with one of the standard (or
			// particularly common) operator or channel symbols.
			// While these shouldn't be any more bad than any
			// arbitrary unusual characters, clients do insist on
			// hardcoding prefixes, even non-standard ones.
			if strings.IndexRune("~&@%+#", int(nick[0])) != -1 {
				return -1e9, os.NewError("Nickname starts with illegal character.")
			}
		}

		return 0, nil
	})

	// Impose the IRC client restriction on idents.
	perm.HookCheckUserData("ident", func(_ string, _, _ *core.User, _, ident string) (int, os.Error) {
		// Do not permit @ or space in an ident.
		if strings.IndexAny(ident, "@ ") != -1 {
			return -1e9, os.NewError("Ident contains @ or space characters.")
		}

		return 0, nil
	})
}
