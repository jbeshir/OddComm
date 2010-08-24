package perm

import "os"

import "oddircd/src/core"


func init() {
	// Add the core permissions for speaking in channels.
	HookChanMsg(true, "", "", banned) 
	HookChanMsg(true, "", "", moderated) 
	HookChanMsg(true, "", "", voiceOverride) 
	HookChanMsg(true, "", "", opOverride) 
}

// If a user has a ban with the mute restriction, they don't get to speak.
func banned(source* core.User, target *core.Channel, msg []byte) (int, os.Error) {
	if Banned(source, target, "mute") {
		return -100, os.NewError("You are banned and cannot speak on the channel.")
	}
	return 0, nil
}

// If a user is affected by speaking being restricted, they don't get to speak.
func moderated(source *core.User, target *core.Channel, msg []byte) (int, os.Error) {
	if Restricted(source, target, "mute") {
		return -100, os.NewError("Channel is moderated. You must have permission to speak on the channel.")
	}
	return 0, nil
}

// Voiced users overrride most restrictions on speaking.
func voiceOverride(source *core.User, target *core.Channel, msg []byte) (int, os.Error) {
	if m := target.GetMember(source); m != nil {
		if m.Data("voiced") != "" {
			return 100, nil
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
