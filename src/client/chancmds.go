package client

import "io"
import "strings"

import "oddcomm/src/core"
import "oddcomm/lib/irc"
import "oddcomm/lib/perm"


func init() {
	var c *irc.Command
	if Commands == nil {
		Commands = irc.NewCommandDispatcher()
	}

	c = new(irc.Command)
	c.Name = "JOIN"
	c.Handler = cmdJoin
	c.Minargs = 1
	c.Maxargs = 1
	Commands.Add(c)

	c = new(irc.Command)
	c.Name = "PART"
	c.Handler = cmdPart
	c.Minargs = 1
	c.Maxargs = 2
	Commands.Add(c)

	c = new(irc.Command)
	c.Name = "KICK"
	c.Handler = cmdKick
	c.Minargs = 2
	c.Maxargs = 3
	Commands.Add(c)

	c = new(irc.Command)
	c.Name = "TOPIC"
	c.Handler = cmdTopic
	c.Minargs = 1
	c.Maxargs = 2
	Commands.Add(c)

	c = new(irc.Command)
	c.Name = "INVITE"
	c.Handler = cmdInvite
	c.Minargs = 2
	c.Maxargs = 2
	Commands.Add(c)
}


func cmdJoin(u *core.User, w io.Writer, params [][]byte) {
	c := w.(*Client)

	chans := strings.Split(string(params[0]), ",", -1)
	for _, channame := range chans {
		if channame[0] == '#' {
			channame = channame[1:]
		}

		ch := core.GetChannel("", channame)
		if ok, err := perm.CheckJoin(u, ch); ok {
			ch.Join([]*core.User{u})
		} else {
			c.WriteTo(nil, "495", "#%s :%s", ch.Name(), err)
		}
	}
}

func cmdPart(u *core.User, w io.Writer, params [][]byte) {
	chans := strings.Split(string(params[0]), ",", -1)
	for _, c := range chans {
		channame := c
		if channame[0] == '#' {
			channame = channame[1:]
		}

		if ch := core.FindChannel("", channame); ch != nil {
			var message string
			if len(params) > 1 {
				message = string(params[1])
			}
			ch.Remove(u, u, message)
		}
	}
}

func cmdKick(u *core.User, w io.Writer, params [][]byte) {
	c := w.(*Client)

	var ch *core.Channel
	channame := string(params[0])
	if channame[0] == '#' {
		channame = channame[1:]
	}
	if ch = core.FindChannel("", channame); ch == nil {
		return
	}

	nicks := strings.Split(string(params[1]), ",", -1)
	for _, nick := range nicks {
		target := core.GetUserByNick(nick)
		if target == nil {
			continue
		}
		if ok, err := perm.CheckRemove(u, target, ch); ok {
			var message string
			if len(params) > 2 {
				message = string(params[2])
			}
			ch.Remove(u, target, message)
		} else {
			c.WriteTo(nil, "482", "#%s :%s", ch.Name(), err)
		}
	}
}

func cmdTopic(u *core.User, w io.Writer, params [][]byte) {
	c := w.(*Client)

	var ch *core.Channel
	if params[0][0] == '#' {
		channame := string(params[0][1:])
		ch = core.FindChannel("", channame)
	}
	if ch == nil {
		c.WriteTo(nil, "403", "%s %s :No such channel.", u.Nick(),
			params[0])
		return
	}

	// If we're displaying the topic...
	if len(params) < 2 {
		topic, setby, setat := ch.GetTopic()
		if topic != "" {
			c.WriteTo(nil, "332", "#%s :%s", ch.Name(), topic)
			c.WriteTo(nil, "333", "#%s %s %s", ch.Name(), setby,
				setat)
		} else {
			c.WriteTo(nil, "331", "#%s :No topic is set.",
				ch.Name())
		}
		return
	}

	// Otherwise, we're setting the topic.
	ch.SetTopic(u, string(params[1]))
}

func cmdInvite(u *core.User, w io.Writer, params [][]byte) {
	c := w.(*Client)

	if len(params[1]) < 2 || params[1][0] != '#' {
		c.WriteTo(nil, "403", "%s :%s", params[1], "No such channel.")
		return
	}

	ch := core.FindChannel("", string(params[1][1:]))
	if ch == nil {
		c.WriteTo(nil, "403", "%s :%s", params[1], "No such channel.")
		return
	}

	targets := strings.Split(string(params[0]), ",", -1)
	for _, t := range targets {
		target := core.GetUserByNick(string(t))
		if target == nil {
			c.WriteTo(nil, "401", "%s :%s", t, "No such nick.")
			continue
		}

		if ok, err := perm.CheckUserMsg(u, target, []byte(ch.Name()),
			"invite"); ok {
			if v := target.Data("away"); v != "" {
				c.WriteTo(nil, "301", "%s :%s",
					target.Nick(), v)
			}
			target.Message(u, []byte(ch.Name()), "invite")
			ch.Message(u, []byte(target.Nick()), "invite")
			c.WriteTo(nil, "341", "%s #%s", target.Nick(),
				ch.Name())
		} else {
			c.WriteTo(nil, "404", "%s :%s", target.Nick(), err)
		}
		continue
	}
}
