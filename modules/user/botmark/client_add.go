package botmark

import "oddircd/src/client"

func init() {
	client.UserModes.AddSimple('B', "bot")
}
