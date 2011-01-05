package ochanctrl

import "oddcomm/src/client"
import "oddcomm/src/core"
import "oddcomm/lib/irc"
import "oddcomm/lib/perm"


// Add command.
func init() {
	c := new(irc.Command)
	c.Name = "OKICK"
	c.Handler = cmdOkick
	c.Minargs = 2
	c.Maxargs = 3
	c.OperFlag = "chanctrl"
	client.Commands.Add(c)
}

func cmdOkick(source interface{}, params [][]byte) {
	c := source.(*client.Client)

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

	perm, err := perm.CheckRemovePerm(c.User(), target, ch)
	if perm < -1000000 {
		c.WriteTo(nil, "482", "#%s :%s", ch.Name(), err)
		return
	}

	var message string
	if len(params) > 2 {
		message = string(params[2])
	}
	ch.Remove(nil, target, message)
}
