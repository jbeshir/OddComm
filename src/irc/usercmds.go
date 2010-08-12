package irc

import "io"
import "fmt"

import "oddircd/src/core"


// QUIT command; removes the user.
func CmdQuit(u *core.User, w io.Writer, params [][]byte) {
	if len(params) > 0 {
		u.Delete(fmt.Sprintf("Quit: %s", params[0]))
	} else {
		u.Delete("Client exited")
	}
}
