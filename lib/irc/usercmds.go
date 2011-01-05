package irc

import "oddcomm/src/core"


// QUIT command; removes the user.
func CmdQuit(source interface{}, params [][]byte) {
	u := source.(*core.User)

	if len(params) > 0 {
		u.Delete(u, string(params[0]))
	} else {
		u.Delete(u, "")
	}
}
