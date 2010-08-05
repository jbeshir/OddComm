package irc

import "fmt"

import "oddircd/core"


func User(u *core.User, params [][]byte) {
	if (u.Data("ident") == "") { return }
	
	u.SetData("ident", string(params[0]))
	u.SetData("realname", string(params[2]))
}

func Nick(u *core.User, params [][]byte) {
	u.SetNick(string(params[0]))
}

func Quit(u *core.User, params [][]byte) {
	if len(params) > 0 {
		u.Remove(fmt.Sprintf("Quit: %s", params[0]))
	} else {
		u.Remove("Client exited")
	}
}
