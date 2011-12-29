package logic

import "oddcomm/src/core/connect"
import "oddcomm/src/core/connect/mmn"


// React to a line received from a node.
// Must be called from the node's goroutine.
func (n *Node) receiveLine(line *mmn.Line) {
	switch true {

		case line.VersionList != nil:
			n.receiveVersionList(line.VersionList.Versions)

		case line.Version != nil:
			n.receiveVersion(*line.Version)

		case line.Cap != nil:
			n.receiveCap(line.Cap.Capabilities)

		case line.Degraded != nil:
			n.receiveDegraded(*line.Degraded)
	}
}

// Receive a version list.
func (n *Node) receiveVersionList(versions []string) {

	// If this is not a new outgoing connection, error.
	if n.conn.State != connect.ConnStateInitialOutgoing {
		n.conn.Close()
		return
	}

	// Pick a version that we support.
	// At present, just "1".
	var picked string
	for _, version := range versions {
		if version == "1" {
			picked = version
		}
	}

	// Check we had a version in common.
	if picked == "" {
		n.conn.Close()
		return
	}

	// Send our picked version.
	n.conn.WriteLine(connect.MakeVersion(picked))

	// Set this connection's protocol version.
	n.conn.Version = picked

	// Move into capability negotiation state.
	n.conn.State = connect.ConnStateCapabilityNegotiationOutgoing
}

func (n *Node) receiveVersion(version string) {

	// If this is not a new incoming connection, error.
	if n.conn.State != connect.ConnStateInitialIncoming {
		n.conn.Close()
		return
	}

	// Verify that we support this version.
	found := false
	for _, supported := range connect.Versions {
		if version == supported {
			found = true
			break
		}
	}

	if !found {
		n.conn.Close()
		return
	}

	// Set this connection's protocol version.
	n.conn.Version = version

	// Send our capabilities list.
	n.conn.WriteLine(connect.MakeCap(connect.Capabilities))

	// Move into capabilities negotiation state.
	n.conn.State = connect.ConnStateCapabilityNegotiationIncoming
}

func (n *Node) receiveCap(capabilities []string) {

	switch n.conn.State {
	case connect.ConnStateCapabilityNegotiationOutgoing:

		// We're expecting a list of supporting capabilities,
		// and need to send back the subset we support.
		var shared []string
		for _, remoteCapability := range capabilities {
			for _, capability := range connect.Capabilities {
				if capability == remoteCapability {
					shared = append(shared, capability)
					break
				}
			}
		}

		// Set this connection's enabled capabilities.
		n.conn.Capabilities = shared

		// Send back the picked capability set.
		n.conn.WriteLine(connect.MakeCap(shared))

		// Send our degraded notification.
		n.conn.WriteLine(connect.MakeDegraded(Degraded))

		// Move into degraded notification state.
		n.conn.State = connect.ConnStateDegradedNotification

	case connect.ConnStateCapabilityNegotiationIncoming:

		// We're expecting a list of selected capabilities.
		// Verify that we have the capabilities they send.
		for _, selected := range capabilities {
			matched := false
			for _, capability := range connect.Capabilities {
				if capability == selected {
					matched = true
					break
				}
			}

			if !matched {
				// Invalid capability.
				n.conn.Close()
				return
			}
		}

		// Set this connection's enabled capabilities.
		n.conn.Capabilities = capabilities

		// Send our degraded notification.
		n.conn.WriteLine(connect.MakeDegraded(Degraded))

		// Move into degraded notification state.
		n.conn.State = connect.ConnStateDegradedNotification

	default:

		// Invalid state for this message.
		n.conn.Close()
	}
}

func (n *Node) receiveDegraded(nodeDegraded bool) {

	// If we're not in degraded notification state, error.
	if n.conn.State != connect.ConnStateDegradedNotification {
		n.conn.Close()
		return
	}

	// If both us and the other node are degraded, drop the connection.
	if nodeDegraded && Degraded {
		n.conn.Close()
		return
	}

	// SEND CHANGE NONCE

	// Move into synchronisation state.
	n.conn.State = connect.ConnStateSynchronization
}
