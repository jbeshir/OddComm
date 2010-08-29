/*
	The main package for the server.

	Starts every package goroutine, then waits for them to shut down
	before terminating.
*/
package main

import "oddcomm/src/core"
import "oddcomm/src/client"

import modules_user_botmark "oddcomm/modules/user/botmark"
import modules_client_extbans "oddcomm/modules/client/extbans"
import modules_client_login "oddcomm/modules/client/login"
import modules_client_ochanctrl "oddcomm/modules/client/ochanctrl"
import modules_client_opermode "oddcomm/modules/client/opermode"
import modules_client_opflags "oddcomm/modules/client/opflags"
import modules_oper_account "oddcomm/modules/oper/account"
import modules_oper_pmoverride "oddcomm/modules/oper/pmoverride"
import modules_dev_catserv "oddcomm/modules/dev/catserv"
import modules_dev_horde "oddcomm/modules/dev/horde"
import modules_dev_testaccount "oddcomm/modules/dev/testaccount"
import modules_dev_tmmode "oddcomm/modules/dev/tmmode"

func main() {
	// Makes modules be permitted to link in.
	_ = modules_user_botmark.MODULENAME
	_ = modules_client_extbans.MODULENAME
	_ = modules_client_login.MODULENAME
	_ = modules_client_ochanctrl.MODULENAME
	_ = modules_client_opermode.MODULENAME
	_ = modules_client_opflags.MODULENAME
	_ = modules_oper_account.MODULENAME
	_ = modules_oper_pmoverride.MODULENAME
	_ = modules_dev_catserv.MODULENAME
	_ = modules_dev_horde.MODULENAME
	_ = modules_dev_testaccount.MODULENAME
	_ = modules_dev_tmmode.MODULENAME

	var exitList [1]chan int
	var msg chan string
	var exit chan int

	// Start client subsystem.
	msg, exit = client.Start()
	if msg != nil {
		core.AddPackage("oddcomm/src/client", msg)
	}
	exitList[0] = exit

	// Run start hooks.
	core.RunStartHooks()

	// Wait until every package goroutine returns.
	for i := range exitList {
		if exitList[i] != nil {
			<-exitList[i]
		}
	}
}
