/*
	
*/
package perm


var defaultBan = "join mute nick"
var defaultUnrestrict = "join"

// DefaultBan returns the list of restrictions for a default ban.
// They are space-separated.
func DefaultBan() string {
	return defaultBan
}

// DefaultUnrestrict returns the list of restrictions to lift with a default
// unrestriction. They are space separated.
func DefaultUnrestrict() string {
	return defaultUnrestrict
}

// DefaultBanAdd adds the given restriction to the default ban restrictions.
// It may be used only during init.
func DefaultBanAdd(name string) {
	defaultBan += " " + name
}

// DefaultUnrestrictAdd adds the given restriction to the restrictions a
// default unrestriction will lift. It may be used only during init.
func DefaultUnrestrictAdd(name string) {
	defaultUnrestrict += " " + name
}
