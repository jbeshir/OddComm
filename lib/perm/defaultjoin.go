package perm

import "os"

import "oddcomm/src/core"


// If a user has a ban with the join restriction, they don't get to join.
func joinBanned(source* core.User, target *core.Channel) (int, os.Error) {
	if Banned(source, target, "join") {
		return -10000, os.NewError("You are banned and cannot join the channel.")
	}
	return 0, nil
}

// If a user is affected by joining being restricted, they don't get to speak.
func inviteOnly(source *core.User, target *core.Channel) (int, os.Error) {
	if Restricted(source, target, "join") {
		return -100, os.NewError("Channel is invite-only. You must be invited or otherwise have permission to join.")
	}
	return 0, nil
}

// Users can remove themselves by default.
func selfOverride(source, target *core.User, ch *core.Channel) (int, os.Error) {
	if source == target {
		return 100, nil
	}
	return 0, nil
}

// Channel operators are immune to being removed by anything below server op
// level aside themselves. They must be deopped first.
func opKickImmune(source, target *core.User, ch *core.Channel) (int, os.Error) {
	if source == target {
		return 0, nil
	}
	if m := ch.GetMember(target); m != nil {
		if m.Data("op") != "" {
			return -1000000, os.NewError("Target user is an operator, and cannot be kicked without being deopped.")
		}
	}
	return 0, nil
}

// Channel operators with the "ban" flag can remove other users, too.
func opKickOverride(source, target *core.User, ch *core.Channel) (int, os.Error) {
	if HasOpFlag(source, ch, "ban") {
		return 10000, nil
	}
	return 0, nil
}
