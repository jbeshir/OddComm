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


var me string = "modules/dev/horde"


func init() {
	// Join the server on startup.
	core.HookStart(addHorde)
}

func addHorde() {
	src := rand.NewSource(time.Nanoseconds())
	rng := rand.New(src)
	horde := make([]*core.User, 20000)

	// Add the horde.
	for i, _ := range horde {
		data := make([]core.DataChange, 4)
		data[0].Name, data[0].Data = "ip", fmt.Sprintf("%d.%d.%d.%d", rng.Int()%255, rng.Int()%255, rng.Int()%255, rng.Int()%255)
		data[1].Name, data[1].Data = "hostname", fmt.Sprintf("%d.Horde.FakeUsers.PsuedoUserUnion.org", rng.Int()%1000000)
		data[2].Name, data[2].Data = "ident", fmt.Sprintf("horde-%d", rng.Int()%1000000)
		data[3].Name, data[3].Data = "account", fmt.Sprintf("horde-%d", rng.Int()%1000000)

		horde[i] = core.NewUser("oddcomm/modules/dev/horde", nil, true, "", data)
		horde[i].SetNick(me, fmt.Sprintf("horde-%d", rng.Int()%1000000))
		horde[i].PermitRegistration(me)
	}

	// Make a huge channel containing the entire horde.
	core.GetChannel("", "huge").Join(me, horde)

	// Make 100 channels containing roughly a twentieth of the horde each.
	// Each horde user is in an average of roughly five.
	for i := 0; i < 100; i++ {
		joiners := make([]*core.User, len(horde)/20)
		for i, _ := range joiners {
			joiners[i] = horde[rand.Int()%len(horde)]
		}
		name := fmt.Sprintf("big_%d", i)
		core.GetChannel("", name).Join(me, joiners)
	}

	// Make 2000 channels containing roughly 1/400th of the horde each.
	// Each horde user is in an average of roughly five.
	for i := 0; i < 2000; i++ {
		joiners := make([]*core.User, len(horde)/400)
		for i, _ := range joiners {
			joiners[i] = horde[rand.Int()%len(horde)]
		}
		name := fmt.Sprintf("medium_%d", i)
		core.GetChannel("", name).Join(me, joiners)
	}

	// Make horde*2 channels containing roughly four of the horde each.
	// Each horde user is in an average of roughly eight.
	for i := 0; i < len(horde)*2; i++ {
		joiners := make([]*core.User, 4)
		for i, _ := range joiners {
			joiners[i] = horde[rand.Int()%len(horde)]
		}
		name := fmt.Sprintf("small_%d", i)
		core.GetChannel("", name).Join(me, joiners)
	}
}
