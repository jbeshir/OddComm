package irc

import "oddircd/src/core"


// GetIdent gets an ident for a user, substituting "." if none exists.
func GetIdent(u *core.User) (ident string) {
	ident = u.Data("ident")
	if ident == "" {
		ident = "-"
	}
	return
}

// GetHostname gets a hostname for a user, substituting the server name if none
// exists.
func GetHostname(u *core.User) (hostname string) {
	hostname = u.Data("hostname")
	if hostname == "" {
		hostname = "Server.name"
	}
	return
}
