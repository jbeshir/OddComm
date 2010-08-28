package botmark

import "oddcomm/src/client"

func init() {
	client.UserModes.AddSimple('B', "bot")
}
