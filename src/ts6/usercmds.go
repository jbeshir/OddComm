package ts6

import "strconv"

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

	// Add QUIT command.
	c = new(irc.Command)
	c.Name = "QUIT"
	c.Handler = cmdQuit
	c.Minargs = 1
	c.Maxargs = 1
	commands.Add(c)

	// Add KILL command.
	c = new(irc.Command)
	c.Name = "KILL"
	c.Handler = cmdKill
	c.Minargs = 2
	c.Maxargs = 2
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
		// Bad UID, kill.
		s.local.SendLine(nil, uid, "KILL", ":Bad UID")
		return
	}

	// Set the data to be applied to the new user.
	data := make([]core.DataChange, 4)
	data[0].Name, data[0].Data = "ident", string(params[4])
	data[1].Name, data[1].Data = "hostname", string(params[5])
	data[2].Name, data[2].Data = "ip", string(params[6])
	data[3].Name, data[3].Data = "realname", string(params[8])

	// Add the user.
	u := core.NewUser(me, s, true, uid, data)
	if u == nil {
		// Duplicate UID, kill.
		s.local.SendLine(nil, uid, "KILL", ":Duplicate UID")
		return
	}

	// Try to set their nick.
	ts, _ := strconv.Atoi64(string(params[2]))
	setNick(u, string(params[0]), ts)

	// They are now registered.
	u.PermitRegistration(me)
}

// User nick change command.
// May only come from a user.
func cmdNick(source interface{}, params [][]byte) {
	u, ok := source.(*core.User)
	if !ok {
		return
	}

	setNick(u, string(params[0]), -1)
}

// User quit command.
// May only come from a user.
func cmdQuit(source interface{}, params [][]byte) {
	u, ok := source.(*core.User)
	if !ok {
		return
	}

	// Delete the user in question.
	u.Delete(me, u, string(params[0]))
}

// User kill command.
func cmdKill(source interface{}, params [][]byte) {

	// u is allowed to be nil here; it means a kill from a server,
	// which will be translated into a kill from this server.
	u, _ := source.(*core.User)

	// Look up the target.
	target := core.GetUser(string(params[0]))
	if target == nil {
		// No target. Legal- they could be already killed.
		return
	}

	// Delete the user in question.
	target.Delete(me, u, string(params[1]))
}


// Message command.
func cmdPrivmsg(source interface{}, params [][]byte) {

	// u is allowed to be nil here; it means a message from a server,
	// which will be translated into a message from this server.
	u, _ := source.(*core.User)
	t := string(params[0])

	if target := core.GetUser(t); target != nil {
		target.Message(me, u, params[1], "")
		return
	}

	if t[0] == '#' {
		channame := t[1:]
		ch := core.FindChannel("", channame)
		if ch != nil {
			ch.Message(me, u, params[1], "")
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
		target.Message(me, u, params[1], "noreply")
		return
	}

	if t[0] == '#' {
		channame := t[1:]
		ch := core.FindChannel("", channame)
		if ch != nil {
			ch.Message(me, u, params[1], "noreply")
			return
		}
	}
}


// Sets the given user's nick as specified, handling collisions.
// If ts is not -1, it specifies the user's nick timestamp.
// Returns whether the user managed to avoid being killed in a collision.
func setNick(u *core.User, nick string, ts int64) bool {

	// Set their nick.
	// Keep trying until we succeed or lose in a collision.
	for u.SetNick(me, nick, ts) != nil {

		// If we can get the colliding user...
		if col := core.GetUserByNick(nick); col != nil {

			// Get the nick timestamps.
			time1 := ts
			time2 := col.NickTS()

			// If the two users have the same IP, reverse the
			// timestamps; the new one is likely to be replacing
			// the old.
			if u.Data("ip") == col.Data("ip") {
				time1, time2 = time2, time1
			}

			if time1 < time2 {
				// New user loses.
				return collideUser(u)
			}

			if time2 < time1 {
				// Existing user loses.
				collideUser(col)
				continue
			}

			// Everyone loses!
			collideUser(col)
			return collideUser(u)
		}
	}

	return true
}

// Collides the given user so they are no longer using their current nick.
// This can forcibly change their nick, or kill them, depending on support.
// Returns whether the user is still alive; false means they were killed.
func collideUser(u *core.User) bool {

	// At present, just a lazy kill.
	if u.Registered() || u.Owner() == me {
		all(func(l *local) {
			l.SendLine(nil, u.ID(), "KILL", ":Nick Collision")
		})
	}
	u.Delete(me, nil, "Nick Collision")

	return true
}
