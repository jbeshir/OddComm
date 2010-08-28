/*
	Adds metadata to simply mark the user as a bot.
*/
package botmark

import "os"

import "oddcomm/src/core"
import "oddcomm/lib/perm"


var MODULENAME string = "modules/user/botmark"

func init() {
	// Users can mark themselves as a bot.
	perm.HookCheckUserData("bot", func(source, target *core.User, name, value string) (int, os.Error) {
		if source == target {
			return 100, nil
		}
		return 0, nil
	})
}
