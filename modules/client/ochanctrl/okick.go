package ochanctrl

import "io"

import "oddcomm/src/client"
import "oddcomm/src/core"
import "oddcomm/lib/irc"


// Add command.
func init() {
	c := new(irc.Command)
	c.Handler = cmdOkick
	c.Minargs = 2
	c.Maxargs = 3
	client.Commands.Add("OKICK", c)
}

func cmdOkick(u *core.User, w io.Writer, params [][]byte) {
	var ch *core.Channel
	var target *core.User

	channame := string(params[0])
	if channame[0] == '#' {
		channame = channame[1:]
	}
	if ch = core.FindChannel("", channame); ch == nil {
		return
	}

	if target = core.GetUserByNick(string(params[1])); target == nil {
		return
	}

	var message string
	if len(params) > 2 {
		message = string(params[2])
	}
	ch.Remove(nil, target, message)
}
