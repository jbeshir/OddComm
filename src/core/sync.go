package core

import "sync"

// Counts the number of current synchronisers.
var syncCount int
var syncMutex sync.Mutex

// Sync permits all users and channels to be enumerated to get a valid image of
// the server's state, which can be kept up to date by hooks on later changes.
// No new users or channels will be added while synchronisation is ongoing.
//
// The first function given will be called for each user, then the second for
// each channel. They will be called twice. The first time, the second
// parameter will be false, and the second time, it will be true.
//
// Finally, the third function is called twice to permit global state to be
// synchronised, with the given parameter false the first time and true the
// second.
//
// For all of them, during the first call, changes to that piece of state will
// not happen, and an image of it can be taken. The second time notifies that
// all hooks for changes made prior to the first call have been made, and
// future hook calls relating to that piece of state should be taken as changes
// made after the image was taken.
//
// During their first call for a given piece of state, these functions may only
// read from it, not write to it. This would result in a hang, as writes are
// blocked until it returns.
//
// Concurrent calls to this are fully supported.
func Sync(uf func(*User, bool), cf func(*Channel, bool), gf func(bool)) {
	incSync()

	uit := users.Iterate()
	if uit != nil {
		for {
			_, p := uit.Value()
			u := (*User)(p)
			u.mutex.Lock()

			uf(u, false)
			hookRunner <- func() {
				uf(u, true)
			}

			u.mutex.Unlock()
			if !uit.Next() {
				break
			}
		}
	}

	cit := channels.Iterate()
	if cit != nil {
		for {
			_, p := cit.Value()
			ch := (*Channel)(p)
			ch.mutex.Lock()

			cf(ch, false)

			hookRunner <- func() {
				cf(ch, true)
			}

			ch.mutex.Unlock()
			if !cit.Next() {
				break
			}
		}
	}

	global.mutex.Lock()
	gf(false)
	hookRunner <- func() {
		gf(true)
	}
	global.mutex.Unlock()

	decSync()
}

// Increment the number of ongoing synchronisations.
// Locks user and channel addition/removal if this is the first.
func incSync() {
	syncMutex.Lock()
	if syncCount == 0 {
		userMutex.Lock()
		chanMutex.Lock()
	}
	syncCount++
	syncMutex.Unlock()
}

// Decrement the number of ongoing synchronisations.
// Unlocks user and channel addition/removal if this is the last.
func decSync() {
	syncMutex.Lock()
	syncCount--
	if syncCount == 0 {
		chanMutex.Unlock()
		userMutex.Unlock()
	}
	syncMutex.Unlock()
}
