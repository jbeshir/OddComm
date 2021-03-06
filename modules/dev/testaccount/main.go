/*
	Provides password authentication to a test account, with a fixed
	account name and password.
*/
package testaccount

import "os"
import "strings"

import "oddcomm/src/core"
import "oddcomm/lib/perm"


func init() {
	perm.HookCheckLogin(func(_ string, u *core.User, account, authtype, auth string) (int, os.Error) {
		if strings.ToUpper(account) != "TESTACCOUNT" {
			return 0, nil
		}
		if authtype != "password" || auth != "supertest" {
			return 0, nil
		}
		return 100, os.NewError("TestAccount")
	})
}
