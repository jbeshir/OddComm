package ts6

import "oddcomm/src/core"


// Propagate the given line in this direction if the source is not in this
// direction and the server has finished bursting.
// Must be used from a global data hook.
func (l *local) propagate_global(source *server, line string) {
	if source != nil && source.local == l {
		return
	}

	l.mutex.Lock()

	if l.burst_sent == true && !l.disconnecting {
		l.c.Write([]byte(line))
	}

	l.mutex.Unlock()
}


// Propagate the given line in this direction if the source is not in this
// direction and the server has burst this user.
// Must be used from a user data hook.
func (l *local) propagate_user(source *server, u *core.User, line string) {
	if source != nil && source.local == l {
		return
	}

	l.mutex.Lock()

	if l.burst_sent == true && !l.disconnecting {
		l.c.Write([]byte(line))
	}

	l.mutex.Unlock()
}

// Propagate the given line in this direction if the source is not in this
// direction and the server has finished bursting.
// If it could not be propagated due to incomplete burst,
// appends it to the message queue.
func (l *local) propagate_msg(source *server, u *core.User, line string) {
	if source != nil && source.local == l {
		return
	}

	l.mutex.Lock()

	if l.burst_sent == true && !l.disconnecting {
		l.c.Write([]byte(line))
	}

	l.mutex.Unlock()
}

// Propagate the given line in this direction if the source is not in this
// direction and the server has burst this channel.
// Must be used from a channel data hook.
func (l *local) propagate_chan(source *server, ch *core.Channel, line string) {
	if source != nil && source.local == l {
		return
	}

	l.mutex.Lock()

	if l.burst_sent == true && !l.disconnecting {
		l.c.Write([]byte(line))
	}

	l.mutex.Unlock()
}
