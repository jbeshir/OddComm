/*
	Makes operators with the "msg" flag (default) inherently override
	restrictions when sending a private message to a user.
*/
package pmoverride

import "os"

import "oddircd/src/core"
import "oddircd/src/perm"


var MODULENAME string = "modules/oper/pmoverride"

func init() {
	perm.HookUserMsg(true, "", pmOverride)
}

func pmOverride(source *core.User, target *core.User, msg []byte) (int, os.Error) {
	if perm.HasOpFlag(source, nil, "msg") {
		return 1000000, nil
	}
	return 0, nil
}
