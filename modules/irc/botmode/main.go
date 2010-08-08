/*
	Adds a user mode that simply marks the user as a bot.
*/
package botmode

import "oddircd/src/client"


var MODULENAME string = "modules/irc/botmode"

func init() {
	client.UserModes.AddSimple('B', "bot")
}
