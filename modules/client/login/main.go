/*
	Permits logging into accounts via configuration-defined passwords.
*/
package login

import "io"
import "strings"

import "oddircd/src/client"
import "oddircd/src/core"
import "oddircd/src/irc"


var MODULENAME = "modules/client/account"

var passwords map[string]string
var names map[string]string

func init() {
	passwords = make(map[string]string)
	names = make(map[string]string)

	// Add login command.
	c := new(irc.Command)
	c.Handler = cmdLogin
	c.Minargs = 1
	c.Maxargs = 2
	client.Commands.Add("LOGIN", c)

	// We would load from config here.
	passwords["NAMEGDUF"] = "supertest"
	names["NAMEGDUF"] = "Namegduf"
}

func cmdLogin(u *core.User, w io.Writer, params [][]byte) {
	var account string
	var password string

	// If we only got one parameter, use their nick as the account name.
	if len(params) == 1 {
		account = u.Nick()
		password = string(params[0])
	} else {
		account = string(params[0])
		password = string(params[1])
	}

	// To uppercase!
	account = strings.ToUpper(account)

	// Try to log them in.
	if v, ok := passwords[account]; ok && v == password {
		u.SetData(nil, "account", names[account])
	} else {
		u.Message(nil, []byte("Invalid username or password."),
		          "noreply")
	}
}
