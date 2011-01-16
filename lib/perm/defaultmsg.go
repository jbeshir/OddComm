package perm

import "os"

import "oddcomm/src/core"


// If a user is not in a channel, they don't get to invite users into it.
func externalInvite(_ string, source, target *core.User, msg []byte) (int, os.Error) {
	if source == nil {
		return 0, nil
	}

	var ch *core.Channel
	if ch = core.FindChannel("", string(msg)); ch == nil {
		return -100, os.NewError("No such channel.")
	}
	if m := ch.GetMember(source); m == nil {
		return -100, os.NewError("You are not in that channel.")
	}
	return 0, nil
}

// If the target is already in a channel, you can't invite them.
// This is deliberately weaker than externalInvite, so external users cannot
// use invites to see whether a user is in the channel.
func stupidInvite(_ string, source, target *core.User, msg []byte) (int, os.Error) {
	var ch *core.Channel
	if ch = core.FindChannel("", string(msg)); ch == nil {
		return -100, os.NewError("No such channel.")
	}
	if m := ch.GetMember(target); m != nil {
		return -99, os.NewError("The target is already in that channel.")
	}
	return 0, nil
}


// If a user is not in the channel, they don't get to message it.
func externalMsg(_ string, source *core.User, target *core.Channel, msg []byte) (int, os.Error) {
	if m := target.GetMember(source); m == nil {
		return -100, os.NewError("You are not in the channel.")
	}
	return 0, nil
}

// If a user has a ban with the mute restriction, they don't get to speak.
func muteBanned(_ string, source *core.User, target *core.Channel, msg []byte) (int, os.Error) {
	if Banned(source, target, "mute") {
		return -10000, os.NewError("You are banned and cannot speak on the channel.")
	}
	return 0, nil
}

// If a user is affected by speaking being restricted, they don't get to speak.
func moderated(_ string, source *core.User, target *core.Channel, msg []byte) (int, os.Error) {
	if Restricted(source, target, "mute") {
		return -100, os.NewError("Channel is moderated. You must have permission to speak on the channel.")
	}
	return 0, nil
}

// Voiced users overrride most restrictions on speaking.
func voiceOverride(_ string, source *core.User, target *core.Channel, msg []byte) (int, os.Error) {
	if m := target.GetMember(source); m != nil {
		if m.Data("voiced") != "" {
			return 100, nil
		}
	}
	return 0, nil
}

// Channel operators with the "msg" flag override most restrictions on speaking.
func opMsgOverride(_ string, source *core.User, target *core.Channel, msg []byte) (int, os.Error) {
	if HasOpFlag(source, target, "msg") {
		return 10000, nil
	}
	return 0, nil
}
