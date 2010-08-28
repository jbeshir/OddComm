package ochanctrl

import "io"

import "oddcomm/src/client"
import "oddcomm/src/core"
import "oddcomm/lib/irc"
import "oddcomm/lib/perm"


// Add command.
func init() {
	c := new(irc.Command)
	c.Name = "OJOIN"; c.Handler = cmdOjoin
	c.Minargs = 1; c.Maxargs = 1
	c.OperFlag = "chanctrl"
	client.Commands.Add(c)
}

func cmdOjoin(u *core.User, w io.Writer, params [][]byte) {
	c := w.(*client.Client)

	channame := string(params[0])
	if channame[0] == '#' {
		channame = channame[1:]
	}
	
	ch := core.GetChannel("", channame)
	if perm, err := perm.CheckJoinPerm(u, ch); perm < -1000000 {
		c.WriteTo(nil, "495", "#%s :%s", ch.Name(), err)
		return
	}

	ch.Join(u)
	if m := ch.GetMember(u); m != nil {
		m.SetData(nil, "serverop", "on")
	}
}
