package main

/*
import "fmt"
import "os/signal"

import "oddcomm/src/core"


// Signal handler.
func init() {
	go func() {
		for s := range signal.Incoming {
			if sig, ok := s.(signal.UnixSignal); ok {

				// SIGTERM or SIGINT request a shutdown.
				if sig == 15 || sig == 2 {
					fmt.Printf("Received signal %s, terminating.\n", sig)
					core.Shutdown()
				}
			}
		}
	}()
}
*/
