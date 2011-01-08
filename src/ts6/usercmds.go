package ts6

import "oddcomm/src/core"
import "oddcomm/lib/irc"


func init() {
	var c *irc.Command

	// Add UID command.
	c = new(irc.Command)
	c.Name = "UID"
	c.Handler = cmdUid
	c.Minargs = 9
	c.Maxargs = 9
	commands.Add(c)

	// Add PRIVMSG command.
	c = new(irc.Command)
	c.Name = "PRIVMSG"
	c.Handler = cmdPrivmsg
	c.Minargs = 2
	c.Maxargs = 2
	commands.Add(c)

	// Add NOTICE command.
	c = new(irc.Command)
	c.Name = "NOTICE"
	c.Handler = cmdNotice
	c.Minargs = 2
	c.Maxargs = 2
	commands.Add(c)
}


// User introduction command.
// May only be from a server, indicating a new user on that server.
func cmdUid(source interface{}, params [][]byte) {
	s, ok := source.(*server)
	if !ok {
		return
	}

	// Servers can only introduce users whose UID begins with their SID.
	uid := string(params[7])
	if len(uid) != 9 || uid[:3] != s.sid {
		return
	}

	// Set the data to be applied to the new user.
	data := make([]core.DataChange, 5)
	data[0].Name, data[0].Data = "nickts", string(params[2])
	data[1].Name, data[1].Data = "ident", string(params[4])
	data[2].Name, data[2].Data = "hostname", string(params[5])
	data[3].Name, data[3].Data = "ip", string(params[6])
	data[4].Name, data[4].Data = "realname", string(params[8])

	// Add the user, set their nick, and register.
	u := core.NewUser("oddcomm/src/ts6", s, true, uid, data)
	if u == nil {
		// Duplicate UID!
		return
	}

	// Set their nick.
	if u.SetNick(string(params[0])) != nil {
		u.SetNick("")
	}

	// They are now registered.
	u.PermitRegistration()
}


// Message command.
func cmdPrivmsg(source interface{}, params [][]byte) {

	// u is allowed to be nil here; it means a message from a server, which
	// will be translated into a message from this server.
	u, _ := source.(*core.User)
	t := string(params[0])

	if target := core.GetUser(t); target != nil {
		target.Message(u, params[1], "")
		return
	}

	if t[0] == '#' {
		channame := t[1:]
		ch := core.FindChannel("", channame)
		if ch != nil {
			ch.Message(u, params[1], "")
			return
		} 
	}
}


// No-reply message (NOTICE) command.
func cmdNotice(source interface{}, params [][]byte) {

	// u is allowed to be nil here; it means a message from a server, which
	// will be translated into a message from this server.
	u, _ := source.(*core.User)
	t := string(params[0])

	if target := core.GetUser(t); target != nil {
		target.Message(u, params[1], "noreply")
		return
	}

	if t[0] == '#' {
		channame := t[1:]
		ch := core.FindChannel("", channame)
		if ch != nil {
			ch.Message(u, params[1], "noreply")
			return
		} 
	}
}
