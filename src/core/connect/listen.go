package connect

import "crypto/tls"

// Listen for incoming node connections, which are sent on the given channel.
// RemoteCerts must be setup by this point.
func Listen(addr string, ch chan<- *tls.Conn) {

	config := new(tls.Config)
	config.Certificates = []tls.Certificate{ Cert }
	config.AuthenticateClient = true

	listener, err := tls.Listen("tcp", addr, config)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		tlsConn := conn.(*tls.Conn)

		err = tlsConn.Handshake()
		if err != nil {
			println(err.Error())
			tlsConn.Close()
			continue
		}

		ch <- tlsConn
	}
}
