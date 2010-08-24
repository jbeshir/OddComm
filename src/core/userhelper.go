package core

import "strings"

// GetIdent gets an ident for a user, substituting "." if none exists.
func (u *User) GetIdent() (ident string) {
	ident = u.Data("ident")
	if ident == "" {
		ident = "-"
	}
	return
}

// GetHostname gets a hostname for a user, substituting the server name if none
// exists.
func (u *User) GetHostname() (hostname string) {
	hostname = u.Data("hostname")
	if hostname == "" {
		hostname = "Server.name"
	}
	return
}

// GetSetBy gets a string representing who set a piece of metadata set by the
// user. It returns the user's nick!ident@host if they are logged out, and
// their account name if they are logged in.
func (u *User) GetSetBy() (setby string) {
	wait := make(chan bool)
	corechan <- func() {
		if v := TrieGet(&u.data, "account"); v != nil {
			setby = v.(string)
		} else {
			ident := "-"
			if v := TrieGet(&u.data, "ident"); v != nil {
				ident = v.(string)
			}
			hostname := "Server.name"
			if v := TrieGet(&u.data, "hostname"); v != nil {
				hostname = v.(string)
			}
			setby = u.nick + "!" + ident + "@" + hostname
		}
		wait <- true
	}
	<-wait
	return
}

// GetDecentBan gets a reasonable default ban string on this user, without
// a ban, ban exception, or unrestriction prefix attached.
// At present, it uses an account ban if they are logged into an account, or
// a host ban on their ident@host otherwise. As such, it should be used only
// if a means to adminstrate both kinds of ban is available.
func (u *User) GetDecentBan() string {

	// Logged in? Ban them by account.
	if v := u.Data("account"); v != "" {
		return "account " + v
	}

	ident := u.GetIdent()
	hostname := u.GetHostname()

	if ident[0] == '~' {
		ident = ident[1:]
	}

	if v := strings.IndexRune(hostname, '.'); v != -1 && len(hostname) > v+1 {
		// Decent for all resolved hosts.
		if hostname != u.Data("ip") {
			hostname = "*" + hostname[v+1:]
		} else {
			// Decent for IPv4.
			v = strings.LastIndex(hostname, ".")
			hostname = hostname[0:v+1] + "*"
		}
	} else if v := strings.IndexRune(hostname, ':'); v != -1 && len(hostname) > v+1 {
		// Decent for IPv6. Maybe.
		// It'll do until a better idea shows up.
		v = strings.LastIndex(hostname, ":")
		hostname = hostname[0:v+1] + "*"
	}

	return "host *!*" + ident + "@" + hostname
}
