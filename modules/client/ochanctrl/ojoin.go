package ochanctrl

import "io"

import "oddcomm/src/client"
import "oddcomm/src/core"
import "oddcomm/lib/irc"


// Add command.
func init() {
	c := new(irc.Command)
	c.Name = "OJOIN"; c.Handler = cmdOjoin
	c.Minargs = 1; c.Maxargs = 1
	c.OperFlag = "chanctrl"
	client.Commands.Add(c)
}

func cmdOjoin(u *core.User, w io.Writer, params [][]byte) {
	channame := string(params[0])
	if channame[0] == '#' {
		channame = channame[1:]
	}
	
	ch := core.GetChannel("", channame)
	ch.Join(u)
	if m := ch.GetMember(u); m != nil {
		m.SetData(nil, "serverop", "on")
	}
}
