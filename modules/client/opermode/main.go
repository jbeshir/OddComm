/*
	Adds a server operator membership mode.

	This sets server operator metadata on the channel membership,
	indicating that the user in question is functioning as a server
	operator on that channel at present.
*/
package opermode

import "oddcomm/src/client"


func init() {
	client.ChanModes.AddMembership('A', "serverop")
	client.ChanModes.AddPrefix('!', "serverop", 1000000)
}
