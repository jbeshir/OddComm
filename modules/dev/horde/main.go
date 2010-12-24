/*
	Add a horde (20k) of test users.
	Tests memory efficiency of channels and users without a local socket
	or goroutine.
*/
package horde

import "fmt"
import "rand"
import "time"

import "oddcomm/src/core"


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
	for i := 0; i < 20000; i++ {
		data := make([]core.DataChange, 4)
		data[0].Name, data[0].Data = "ip", fmt.Sprintf("%d.%d.%d.%d", rng.Int()%255, rng.Int()%255, rng.Int()%255, rng.Int()%255)
		data[1].Name, data[1].Data = "hostname", fmt.Sprintf("%d.Horde.FakeUsers.PsuedoUserUnion.org", rng.Int()%1000000)
		data[2].Name, data[2].Data = "ident", fmt.Sprintf("horde-%d", rng.Int()%1000000)
		data[3].Name, data[3].Data = "account", fmt.Sprintf("horde-%d", rng.Int()%1000000)
		data[0].Next, data[1].Next = &data[1], &data[2]
		data[2].Next = &data[3]

		u := core.NewUser("oddcomm/modules/dev/horde", true, "", &data[0])
		u.SetNick(fmt.Sprintf("horde-%d", rng.Int()%1000000))
		u.PermitRegistration()

		// Join 5 random "big" channels.
		// Channel count 100, average size roughly 1000.
		for j := 0; j < 5; j++ {
			name := fmt.Sprintf("big_%d", rng.Int()%100)
			core.GetChannel("", name).Join(u)
		}

		// Join 5 random "medium" channels.
		// Channel count 2000, average size roughly 50.
		for j := 0; j < 5; j++ {
			name := fmt.Sprintf("medium_%d", rng.Int()%2000)
			core.GetChannel("", name).Join(u)
		}

		// Join 10 random "small" channels.
		// Channel count 50000, average size roughly 4.
		for j := 0; j < 10; j++ {
			name := fmt.Sprintf("small_%d", rng.Int()%50000)
			core.GetChannel("", name).Join(u)
		}
	}
}
