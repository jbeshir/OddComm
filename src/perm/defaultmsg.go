package perm

import "os"

import "oddircd/src/core"


func init() {
	// Add the core metadata effects for speaking in channels.
	HookChanMsg(true, "", "", moderated) 
	HookChanMsg(true, "", "", voiceOverride) 
	HookChanMsg(true, "", "", opOverride) 
}

// Moderated being set on a channel prevents speech.
func moderated(source *core.User, target *core.Channel, msg []byte) (int, os.Error) {
	if target.Data("moderated") !=  "" {
		return -100, os.NewError("Channel is moderated. You must have permission to speak on the channel.")
	}
	return 0, nil
}

// Voiced users overrride most restrictions on speaking.
func voiceOverride(source *core.User, target *core.Channel, msg []byte) (int, os.Error) {
	if m := target.GetMember(source); m != nil {
		if m.Data("voiced") != "" {
			return 1000, nil
		}
	}
	return 0, nil
}

// Channel operators override most restrictions on speaking.
func opOverride(source *core.User, target *core.Channel, msg []byte) (int, os.Error) {
	if HasOpFlag(source, target, "msg") {
		return 10000, nil
	}
	return 0, nil
}
