package ochanctrl

import "strings"

import "oddcomm/src/core"
import "oddcomm/src/client"
import "oddcomm/lib/irc"
import "oddcomm/lib/perm"

func init() {
	c := new(irc.Command)
	c.Name = "OMODE"
	c.Handler = cmdOmode
	c.Minargs = 2
	c.Maxargs = 3
	c.OperFlag = "chanctrl"
	client.Commands.Add(c)
}


func cmdOmode(source interface{}, params [][]byte) {
	c := source.(*client.Client)

	channame := string(params[0])
	if channame[0] == '#' {
		channame = channame[1:]
	}
	ch := core.FindChannel("", channame)
	if ch == nil {
		c.SendLineTo(nil, "501", "%s %s :%s", c.User().Nick(), params[0],
			"No such channel.")
		return
	}

	// If we're viewing the modes of a channel...
	if len(params) < 2 {
		modeline := client.ChanModes.GetModes(ch)
		ts := ch.TS()
		c.SendLineTo(nil, "324", "#%s +%s", ch.Name(), modeline)
		c.SendLineTo(nil, "329", "#%s %d", ch.Name(), ts)
		return
	}

	// If we're listing list modes on a channel...
	if params[1][0] != '+' && params[1][0] != '-' {
		var badmodes string
		for _, mode := range string(params[1]) {
			name := client.ChanModes.GetName(mode)
			if name == "" {
				badmodes += string(mode)
				continue
			}

			if perm, err := perm.CheckChanViewDataPerm(me, c.User(), ch, name); perm < -1000000 {
				c.SendLineTo(nil, "482", "#%s :%s", ch.Name(), err)
				continue
			}

			// Different, fixed numerics for different
			// modes. Stupid protocol.
			num := "941"
			endnum := "940"
			switch mode {
			case 'b':
				num = "367"
				endnum = "368"
			case 'e':
				num = "348"
				endnum = "349"
			case 'I':
				num = "346"
				endnum = "347"
			}

			valid := client.ChanModes.ListMode(ch, int(mode),
				func(p, v string) {
					var setTime string = "0"
					var setBy string = "Server.name"
					words := strings.Fields(v)
					for _, word := range words {
						if len(word) > 6 && word[:6] == "setat-" {
							setTime = word[6:]
							continue
						}
						if len(word) > 6 && word[:6] == "setby-" {
							setBy = word[6:]
							continue
						}
					}

					c.SendLineTo(nil, num, "#%s %s %s %s",
						ch.Name(), p, setBy, setTime)
				})
			if valid {
				c.SendLineTo(nil, endnum,
					"#%s :End of mode list.",
					ch.Name())
			} else {
				badmodes += string(mode)
			}
		}
		if badmodes != "" {
			if badmodes != string(params[1]) {
				c.SendLineTo(nil, "501", "Unknown list modes: %s", badmodes)
				return
			}
			// If ALL the mode characters were invalid, we let it
			// fall through and try to treat it as setting modes.
		} else {
			return
		}
		return
	}

	// Otherwise, we're setting modes.
	var mpars []string
	if len(params) == 3 {
		mpars = strings.Fields(string(params[2]))
	}
	changes, err := client.ChanModes.ParseModeLine(c.User(), ch, params[1], mpars)
	if err != nil {
		c.SendLineTo(nil, "501", "%s", err)
	}

	todo := make([]core.DataChange, 0, len(changes))
	for _, cha := range changes {
		if cha.Member != nil {
			num, err := perm.CheckMemberDataPerm(me, c.User(), cha.Member, cha.Name, cha.Data)
			if num < -1000000 {
				c.SendLineTo(nil, "482", "#%s %s: %s", ch.Name(), cha.Name, err)
				continue
			}
		} else {
			num, err := perm.CheckChanDataPerm(me, c.User(), ch, cha.Name, cha.Data)
			if num < -1000000 {
				c.SendLineTo(nil, "482", "#%s %s: %s", ch.Name(), cha.Name, err)
				continue
			}
		}
		todo = append(todo, cha)
	}

	if todo != nil {
		ch.SetDataList(me, nil, todo)
	}

}
