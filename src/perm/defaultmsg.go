package perm

import "os"

import "oddircd/src/core"


func init() {
	// Add the core permissions for speaking in channels.
	HookChanMsg(true, "", "", externalMsg)
	HookChanMsg(true, "", "", muteBanned)
	HookChanMsg(true, "", "", moderated) 
	HookChanMsg(true, "", "", voiceOverride) 
	HookChanMsg(true, "", "", opMsgOverride)
}

// If a user is not in the channel, they don't get to message it.
func externalMsg(source *core.User, target *core.Channel, msg []byte) (int, os.Error) {
	if m := target.GetMember(source); m == nil {
		return -100, os.NewError("You are not in the channel.")
	}
	return 0, nil
}

// If a user has a ban with the mute restriction, they don't get to speak.
func muteBanned(source *core.User, target *core.Channel, msg []byte) (int, os.Error) {
	if Banned(source, target, "mute") {
		return -10000, os.NewError("You are banned and cannot speak on the channel.")
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

// Channel operators with the "msg" flag override most restrictions on speaking.
func opMsgOverride(source *core.User, target *core.Channel, msg []byte) (int, os.Error) {
	if HasOpFlag(source, target, "msg") {
		return 10000, nil
	}
	return 0, nil
}
