package irc

import "fmt"
import "io"

// SendLine sends a formatted line from the given source, to the given target.
func SendLine(w io.Writer, source, target, cmd, format string, args ...interface{}) {
	newargs := make([]interface{}, len(args)+3)
	newargs[0] = source
	newargs[1] = cmd
	newargs[2] = target
	copy(newargs[3:], args)

	fmt.Fprintf(w, ":%s %s %s " + format + "\r\n", newargs...)
}

// SendFrom sends a given prewritten line, prefixd by the given source.
// source may be a nil interface or a nil value, in which case the line will
func SendFrom(w io.Writer, source, format string, args ...interface{}) {
	newargs := make([]interface{}, len(args)+1)
	newargs[0] = source
	copy(newargs[1:], args)

	fmt.Fprintf(w, ":%s " + format + "\r\n", newargs...)
}


