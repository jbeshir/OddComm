/*
	Persistence package.

	Provides loading and saving of global state and configuration,
	handling default values.
*/
package persist

import "encoding/json"
import "errors"
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
func Load(raw io.Reader) (err error) {

	// Get decoder.
	r := json.NewDecoder(raw)

	// Read version.
	var version struct{Version string}
	if err = r.Decode(&version); err != nil {
		return
	}

	// Check the version is supported.
	switch version.Version {
	case "":
		return errors.New("Settings file missing version.")
	case "1":
	default:
		return errors.New("Settings file version unsupported.")
	}

	// Add defaults for settings which must exist.
	requiredState()

	// Load state.
	var data struct{Name, Value string}
	for err == nil {
		err = r.Decode(&data)
		if err == nil {
			if data.Name == "" || data.Value == "" {
				err = errors.New("Settings file corrupt.")
			} else {
				core.Global.SetData("lib/persist", nil,
					data.Name, data.Value)
			}
		}
	}
	if err == io.EOF {
		err = nil
	}

	return
}

// Save saves state to the given writer.
// This state can be fed to Load() in a new instance of the server.
// At present, saved state is all global metadata.
func Save(raw io.Writer) (err error) {

	// Get encoder.
	w := json.NewEncoder(raw)

	// Write version.
	version := struct{Version string}{Version: "1"}
	if err = w.Encode(version); err != nil {
		return
	}

	// Write state.
	core.Sync(nil, nil, func(hook bool) {
		if hook {
			return
		}

		core.Global.DataRange("", func(name, value string) {
			if err != nil {
				return
			}

			var data struct{Name, Value string}
			data.Name = name
			data.Value = value
			err = w.Encode(data)
		})
	})

	return
}

// Adds default values for settings which must exist.
func requiredState() {

	// This should probably be hookable.
	core.Global.SetData("lib/persist", nil, "name", "Server.name")
}
