/*
	Provides server operator "channel control" commands.
	These ignore restrictions (such as not being opped, or being banned)
	below server operator level.

	Commands:

	OJOIN <channel> - Joins the channel.
	OKICK <channel> <nick> - Kicks the given user from the channel.
	OMODE <channel> <modes> [params] - Performs the given mode changes.
*/
package ochanctrl

var MODULENAME string = "modules/client/ochanctrl"
