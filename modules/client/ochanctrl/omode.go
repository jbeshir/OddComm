package ochanctrl

import "io"
import "strings"

import "oddircd/src/core"
import "oddircd/src/client"
import "oddircd/src/irc"
import "oddircd/src/perm"

func init() {
	c := new(irc.Command)
	c.Handler = cmdOmode
	c.Minargs = 2
	c.Maxargs = 42
	client.Commands.Add("OMODE", c)
}


func cmdOmode(u *core.User, w io.Writer, params [][]byte) {
	c := w.(*client.Client)

	channame := string(params[0])
	if channame[0] == '#' {
		channame = channame[1:]
	}
	ch := core.FindChannel("", channame)
	if ch == nil {
		c.WriteTo(nil, "501", "%s %s :%s", u.Nick(), params[0],
		          "No such channel.")
		return
	}

	// If we're viewing the modes of a user or channel...
	if len(params) < 2 {
		modeline := client.ChanModes.GetModes(ch)
		ts := ch.TS()
		c.WriteTo(nil, "324", "#%s +%s", ch.Name(), modeline)
		c.WriteTo(nil, "329", "#%s %d", ch.Name(), ts)
		return
	}

	// If we're listing list modes on a channel...
	if params[1][0] != '+' && params[1][0] != '-' {
		var badmodes string
		for _, mode := range string(params[1]) {
			// Different, fixed numerics for different
			// modes. Stupid protocol.
			num := "941"; endnum := "940"
			switch mode {
			case 'b': num = "367"; endnum = "368"
			case 'e': num = "348"; endnum = "349"
			case 'I': num = "346"; endnum = "347"
			}

			valid := client.ChanModes.ListMode(ch, int(mode),
			                   func(p, v string) {
				var setTime string = "0"
				var setBy string = "Server.name"
				words := strings.Fields(v)
				for _, word := range words {
					if len(word) > 6 && word[0:6] == "setat-" {
						setTime = word[6:]
						continue
					}
					if len(word) > 6 && word[0:6] == "setby-" {
						setBy = word[6:]
						continue
					}
				}

				c.WriteTo(nil, num, "#%s %s %s %s",
				          ch.Name(), p, setBy, setTime)
			})
			if valid {
				c.WriteTo(nil, endnum,
				          "#%s :End of mode list.",
				          ch.Name())
			} else {
				badmodes += string(mode)
			}
		}
		if badmodes != "" {
			if badmodes != string(params[1]) {
				c.WriteTo(nil, "501", "Unknown list modes: %s", badmodes)
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
	changes, err := client.ChanModes.ParseModeLine(u, ch, params[1],
	                                               params[2:])
	if err != nil {
		c.WriteTo(nil, "501", "%s", err)
	}

	prev := &changes
	for cha := changes; cha != nil; cha = cha.Next {
		if cha.Member != nil {
			num, err := perm.CheckMemberDataPerm(u, cha.Member, cha.Name, cha.Data)
			if num < -1000000 {
				c.WriteTo(nil, "482", "#%s %s: %s", ch.Name(), cha.Name, err)
			} else {
				prev = &cha.Next
			}
		} else {
			num, err := perm.CheckChanDataPerm(u, ch, cha.Name, cha.Data)
			if num < -1000000 {
				c.WriteTo(nil, "482", "#%s %s: %s", ch.Name(), cha.Name, err)
				(*prev) = cha.Next
			} else {
				prev = &cha.Next
			}
		}
	}

	if changes != nil {
		ch.SetDataList(nil, changes)
	}

}
