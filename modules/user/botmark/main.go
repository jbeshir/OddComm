/*
	Adds metadata to simply mark the user as a bot.
*/
package botmark

import "os"

import "oddcomm/src/core"
import "oddcomm/lib/perm"


func init() {
	// Users can mark themselves as a bot.
	perm.HookCheckUserData("bot", func(_ string, source, target *core.User, _, _ string) (int, os.Error) {
		if source == target {
			return 100, nil
		}
		return 0, nil
	})
}
