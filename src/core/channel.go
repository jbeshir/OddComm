package core

import "time"


// Represents a channel.
type Channel struct {
	name string
	t string
	ts int64
	data map[string]string
}


// GetChannel returns a channel with the given name and type. Type may be ""
// (for default). If the channel did not previously exist, it is created. If it
// already existed, it is simply returned.
func GetChannel(name string, t string) (ch *Channel) {
	if _, ok := channels[t]; ok {
		if v, ok := channels[t][name]; ok {
			ch = v
		} else {
			ch = new(Channel)
			ch.name = name
			ch.t = t
			ch.ts = time.Seconds()
			channels[t][name] = ch
		}
	} else {
		channels[t] = make(map[string]*Channel)
		ch = new(Channel)
		ch.name = name
		ch.t = t
		ch.ts = time.Seconds()
		channels[t][name] = ch
	}
	return
}

// FindChannel finds a channel with the given name and type, which may be ""
// for the default type. If none exist, it returns nil.
func FindChannel(name string, t string) (ch *Channel) {
	if _, ok := channels[t]; ok {
		if v, ok := channels[t][name]; ok {
			ch = v
		}
	}
	return
}


// Name returns the channel's name.
func (ch *Channel) Name() (name string) {
	wait := make(chan bool)
	corechan <- func() {
		name = ch.name
		wait <- true
	}
	<-wait

	return
}

// SetData sets the given single piece of metadata on the channel.
// Setting it to "" unsets it.
func (ch *Channel) SetData(name string, value string) {
	var oldvalue string

	wait := make(chan bool)
	corechan <- func() {
		oldvalue = ch.data[name]
		if value != "" {
			ch.data[name] = value
		} else {
			ch.data[name] = "", false
		}

		wait <- true
	}
	<-wait

	// runChanDataChangeHooks(u, name, oldvalue, value)

	// c := new(DataChange)
	// c.Name = name
	// c.Data = value
	// runChanDataChangesHooks(ch, c)
}

// SetDataList performs the given list of metadata changes on the channel.
// This is equivalent to lots of SetData calls, except hooks for all data
// changes will receive it as a single list, and it is cheaper.
// There must not be duplicates (changes to the same value) in this list.
func (ch *Channel) SetDataList(c *DataChange) {
	oldvalues := make(map[string]string)
	wait := make(chan bool)
	corechan <- func() {
		for it := c; it != nil; it = c.Next {
			oldvalues[c.Name] = ch.data[c.Name]

			if c.Data != "" {
				ch.data[c.Name] = c.Data
			} else {
				ch.data[c.Name] = "", false
			}
		}

		wait <- true
	}
	<-wait

	// for it := c; it != nil; it = c.Next {
	// 	runChanDataChangeHooks(ch, c.Name, oldvalues[c.Name], c.Data)
	// }
	// runChanDataChangesHooks(ch, c)
}

// Data gets the given piece of metadata.
// If it is not set, this method returns "".
func (ch *Channel) Data(name string) (value string) {
	wait := make(chan bool)
	corechan <- func() {
		value = ch.data[name]
		wait <- true
	}
	<-wait

	return
}

// Message sends a message to the channel.
// source may be nil, indicating a message from the server.
// t may be "" (for default), and indicates the type of message.
func (ch *Channel) Message(source *User, message []byte, t string) {

	// Unregistered users may not send messages.
	if !source.Registered() {
		return
	}

	// We actually just call hooks, and let the subsystems handle it.
	// runMessageHooks(ch, source, message, t)
}

// Delete deletes the channel.
// Remaining users are kicked, if any still exist.
func (ch *Channel) Delete() {
	// This doesn't actually do anything yet.
}
