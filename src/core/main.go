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
	logic.Me = 1

	// Set up nodes.
	var cert *x509.CertPool

	cert = loadCertFile("1.crt")
	logic.NewNode(1, "127.0.0.1:7890", cert)

	cert = loadCertFile("2.crt")
	logic.NewNode(2, "127.0.0.1:7891", cert)

	cert = loadCertFile("3.crt")
	logic.NewNode(3, "127.0.0.1:7892", cert)
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
