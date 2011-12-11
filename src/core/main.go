// Package managing the core state of the server, including coordinating with other OddComm servers.
//
// An API for accessing and setting data is provided here, along with initialising the core with state and settings on startup.
package core

import "crypto/tls"
import "crypto/x509"
import "io/ioutil"

import "oddcomm/src/core/connect"
import "oddcomm/src/core/logic"

func init() {
	var err error

	// Load our TLS certificate.
	connect.Cert, err = tls.LoadX509KeyPair("1.crt", "1.key")
	if err != nil {
		panic(err)
	}

	// Set our node ID.
	logic.Id = 1

	// Set up nodes.
	var info *connect.ConnInfo

	info = new(connect.ConnInfo)
	info.Addr = "127.0.0.1:7890"
	info.Cert = loadCertFile("1.crt")
	logic.NewNode(1, info)

	info = new(connect.ConnInfo)
	info.Addr = "127.0.0.1:7891"
	info.Cert = loadCertFile("2.crt")
	logic.NewNode(2, info)

	info = new(connect.ConnInfo)
	info.Addr = "127.0.0.1:7892"
	info.Cert = loadCertFile("3.crt")
	logic.NewNode(3, info)

	// Start listening for incoming connections.
	// Nodes must be setup first.
	newconns := make(chan *tls.Conn, 10)
	go acceptIncoming(newconns)
	go connect.Listen(newconns)

	// Do initial outgoing connection attempts.

}

// Validates incoming connection client certificates,
// and identifies the node they are associated with,
// then sends the new connection to that node to handle.
func acceptIncoming(ch <-chan *tls.Conn) {
	for {
		conn := <-ch

		state := conn.ConnectionState()
		if len(state.PeerCertificates) == 0 {
			conn.Close()
			continue
		}

		cert := state.PeerCertificates[0]

		// Find the node this connection is from.
		matched := false
		for _, node := range logic.Nodes {
			if node == logic.Me {
				continue
			}

			var verifyOpts x509.VerifyOptions
			verifyOpts.Intermediates = new(x509.CertPool)
			verifyOpts.Roots = node.Cert
			chains, err := cert.Verify(verifyOpts)
			if err != nil {
				continue
			}

			if len(chains) > 0 {
				matched = true
				node.NewConn <- conn
				break
			}
		}

		// No matching node found. Close the connection.
		if !matched {
			conn.Close()
		}
	}
}

func loadCertFile(filename string) *x509.CertPool {

	cert := x509.NewCertPool()
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	ok := cert.AppendCertsFromPEM(file)
	if !ok {
		panic("Unable to parse node certificate file: " + filename)
	}

	return cert
}

