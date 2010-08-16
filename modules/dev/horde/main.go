/*
	Add a massive horde (100k) of test users.
	Tests memory efficiency of users without a local socket.
*/
package horde

import "fmt"
import "rand"
import "time"

import "oddircd/src/core"


// Must be set, must be unique.
var MODULENAME string = "dev/horde"


func init() {
	// Join the server on startup.
	core.HookStart(addHorde)
}

func addHorde() {
	src := rand.NewSource(time.Nanoseconds())
	rng := rand.New(src)

	// Add the horde.
	for i := 0; i < 500; i++ {
		u := core.NewUser("oddircd/modules/dev/horde", true, "")
		u.SetNick(fmt.Sprintf("horde-%d", rng.Int() % 1000000))
		u.SetData(u, "ident", fmt.Sprintf("horde-%d", rng.Int() % 1000000))
		u.SetData(nil, "ip", fmt.Sprintf("%d.%d.%d.%d", rng.Int() % 255, rng.Int() % 255, rng.Int() % 255, rng.Int() % 255))
		u.SetData(nil, "hostname", fmt.Sprintf("%d.Horde.FakeUsers.PsuedoUserUnion.org", rng.Int() % 1000000))
		u.SetData(nil, "account", fmt.Sprintf("horde-%d", rng.Int() % 1000000))

		// Join 5 random "big" channels.
		// Channel count 100, average size 2000.
		for j := 0; j < 5; j++ {
			name := fmt.Sprintf("big_%d", rng.Int() % 100)
			core.GetChannel("", name).Join(u)
		}

		// Join 5 random "medium" channels.
		// Channel count 2000, average size 100.
		for j := 0; j < 5; j++ {
			name := fmt.Sprintf("medium_%d", rng.Int() % 2000)
			core.GetChannel("", name).Join(u)
		}

		// Join 10 random "small" channels.
		// Channel count 100000, average size 10.
		for j := 0; j < 10; j++ {
			name := fmt.Sprintf("small_%d", rng.Int() % 100000)
			core.GetChannel("", name).Join(u)
		}
	}
}
