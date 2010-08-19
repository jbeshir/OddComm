/*
	Implements account-based oper access.
*/
package account

import "strings"

import "oddircd/src/core"

var MODULENAME string = "modules/oper/account"


// Maps accounts to sets of oper flags.
var operAccounts map[string]string

func init() {
	operAccounts = make(map[string]string)

	// Here, we would load things in from the config.
	operAccounts["NAMEGDUF"] = "on"

	// Oper people when they login to their account.
	core.HookUserDataChange("account", func(source, target *core.User, oldvalue, newvalue string) {
		if v, ok := operAccounts[strings.ToUpper(newvalue)]; ok {
			target.SetData(nil, "op", v)
		}
	}, true)
}
