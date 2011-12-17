package connect

import "oddcomm/src/core/connect/mmn"


// List of supported protocol versions.
var Versions = []string { "1" }

// List of supported capabilities.
var Capabilities = []string {}


// Create a VersionList line.
func MakeVersionList() *mmn.Line {

	line := new(mmn.Line)
	line.VersionList = new(mmn.VersionList)

	line.VersionList.Versions = Versions

	return line
}

// Create a Version line.
func MakeVersion(version string) *mmn.Line {

	line := new(mmn.Line)
	line.Version = &version

	return line
}

// Create a Cap line.
func MakeCap(capabilities []string) *mmn.Line {

	line := new(mmn.Line)
	line.Cap = new(mmn.Cap)
	line.Cap.Capabilities = capabilities

	return line
}

// Create a Degraded line.
func MakeDegraded(degraded bool) *mmn.Line {

	line := new(mmn.Line)
	line.Degraded = &degraded

	return line
}
