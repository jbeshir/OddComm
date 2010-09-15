package irc

import "io"

import "oddcomm/src/core"


// QUIT command; removes the user.
func CmdQuit(u *core.User, w io.Writer, params [][]byte) {
	if len(params) > 0 {
		u.Delete(u, string(params[0]))
	} else {
		u.Delete(u, "")
	}
}
