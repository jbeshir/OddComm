/*
	Persistence package.
	
	Provides loading and saving of global state and configuration,
	handling default values.
*/
package persist

import "io"

import "oddcomm/src/core"


// FirstRun adds all initial default state to the core.
// It should be run only if no settings could be loaded, instead of Load().
//
// This adds default values for settings which must exist, and other first run
// defaults.
func FirstRun() {

	// Add defaults for settings which must exist.
	requiredState()

	// This should probably be hookable.
	// STUB: Add defaults.
}

// Load loads state from the given reader.
//
// It will add default values for settings which must exist,
// if not present in the provided state.
//
// State should be in the format produced by an invocation of Save.
// It may be loaded from the same or earlier version of OddComm,
// and is portable across architectures and operating systems.
func Load(r io.Reader) {

	// Add defaults for settings which must exist.
	requiredState()

	// STUB: Add state loading.
}

// Save saves state to the given writer.
// This state can be fed to Load() in a new instance of the server.
//
// At present, saved state is all global metadata.
func Save(w io.Writer) {

	// STUB: Add state saving.
	core.Sync(nil, nil, nil)
}

// Adds default values for settings which must exist.
func requiredState() {

	// This should probably be hookable.
	// STUB: Add defaults.
}
