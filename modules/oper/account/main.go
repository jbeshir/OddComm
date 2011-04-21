/*
	Implements account-based oper access.
*/
package account

import "strings"

import "oddcomm/src/core"


// Maps accounts to sets of oper flags.
var operFlags map[string]string
var operCommands map[string]string
var operType map[string]string

func init() {
	operFlags = make(map[string]string)
	operCommands = make(map[string]string)
	operType = make(map[string]string)

	// Here, we would load things in from the config.
	operType["TESTACCOUNT"] = "Uberdude"
	operFlags["TESTACCOUNT"] = "on"
	operCommands["TESTACCOUNT"] = "OJOIN OMODE DIE"

	// Oper people when they login to their account.
	core.HookUserDataChange("account", func(_ interface{}, source, target *core.User, oldvalue, newvalue string) {
		account := strings.ToUpper(newvalue)
		if v, ok := operType[account]; ok {
			target.SetData(nil, nil, "optype", v)
		}
		if v, ok := operFlags[account]; ok {
			target.SetData(nil, nil, "op", v)
			if v, ok := operCommands[account]; ok {
				target.SetData(nil, nil, "opcommands", v)
			}
		}
	},
		true)
}
