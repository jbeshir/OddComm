/*
	Dummy "service". Demonstrates adding a psuedouser using a module.
*/
package catserv

import "strings"

import "oddcomm/src/core"


// Must be set, must be unique.
var MODULENAME string = "dev/catserv"


var cat *core.User

func init() {
	// Don't show my psuedoclient until I'm done adding myself.
	core.RegistrationHold("oddcomm/modules/dev/catserv")

	// Join the server on startup.
	core.HookStart(addCat)

	core.HookUserNickChange(func(u *core.User, oldnick, newnick string) {
		// If someone was stealing our nick, and they changed nick, try
		// to connect again.
		if u != cat && strings.ToUpper(oldnick) == "CATSERV" {
			addCat()
		}

		// If we've been force nickchanged, disconnect.
		// We will try to reconnect.
		if u == cat && newnick != "CatServ" {
			// Meow!
			cat.Delete(cat, "Meow!")
		}
	},
		true)

	core.HookUserDelete(func(source, u *core.User, _ string) {
		// If we got disconnected or quit, or someone stealing our nick
		// quit, try to reconnect.
		if u == cat || strings.ToUpper(u.Nick()) == "CATSERV" {
			addCat()
		}
	},
		true)

	core.HookUserMessage("", func(source, target *core.User, message []byte) {
		// If someone sent a message to us, say hi.
		// In a real psuedoserver, you could write whatever
		// functionality you liked here.
		if target == cat {
			source.Message(cat, []byte("Meow!"), "noreply")
		}
	})
}

func addCat() {
	// Add CatServ!
	cat = core.NewUser("oddcomm/modules/catserv", true, "")

	// If I suffer a collision, quit for now; I will return when they go
	// away.
	if err := cat.SetNick("CatServ"); err != nil {
		cat.Delete(nil, err.String())
		return
	}

	// Finish registering!
	cat.PermitRegistration()
}
