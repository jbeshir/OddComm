// Provides the logic used to implement the Paxos algorithm for managing writes to state.
//
// Also handles change propagation to other nodes and other aspects of the protocol.
package logic


// Whether this node is currently in a degraded state or not.
var Degraded bool
