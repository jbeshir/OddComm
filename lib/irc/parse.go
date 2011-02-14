package irc

import "bytes"
import "io"

// Read line reads lines from the input, using the given byte buffer, and calls the
// given function with each line. When an error occurs, it returns an error message
// explaining it.
func ReadLine(r io.Reader, b []byte, f func(line []byte)) (errMsg string) {
	var count int
	for {
		// If we have no room in our input buffer to read, the user
		// has overrun their input buffer.
		if count == cap(b) {
			errMsg = "Input Buffer Exceeded"
			break
		}

		// Try to read from the user.
		n, err := r.Read(b[count:cap(b)])
		if err != nil {
			// This happens if the user is disconnected
			// by other code. In this case, the error message
			// will be ignored.
			errMsg = err.String()
			break
		}
		count += n
		b = b[:count]

		for {
			// Search for an end of line, then keep going until we
			// stop finding eol characters, to eat as many as
			// possible in the same operation.
			eol := -1
			for i := range b {
				if b[i] == '\r' || b[i] == '\n' || b[i] == 0 {
					eol = i
				} else if eol != -1 {
					break
				}
			}

			// If we didn't find one, wait for more input.
			if eol == -1 {
				break
			}

			// Get the line, with no line endings.
			line := b[:eol]
			end := len(line)
			for end > 0 {
				endchar := line[end-1]
				if endchar == '\r' || endchar == '\n' {
					end--
				} else {
					break
				}
			}
			if end != len(line) {
				line = line[:end]
			}

			// Ignore blank lines.
			if len(line) == 0 {
				if len(b)-eol-1 >= 0 {
					b = b[:len(b)-eol-1]
					continue
				} else {
					b = b[:0]
					break
				}
			}

			// Run the function.
			f(line)

			// If we have remaining input for the next line, move
			// it down and cut the buffer to it.
			// Otherwise, clear it.
			if len(b)-eol-1 >= 0 {
				copy(b, b[eol+1:])
				b = b[:len(b)-eol-1]
			} else {
				b = b[:0]
				break
			}
		}

		count = len(b)
	}
	return
}

func Parse(d CommandDispatcher, line []byte, regged bool) (origin []byte, command *Command, params [][]byte, err *ParseError) {

	// We can handle up to 50 parameters. This is plenty.
	var param_array [50][]byte
	params = param_array[:0]
	var word []byte

	// Define function for moving to the next word.
	var nextword = func() {
		space := bytes.IndexByte(line, ' ')
		if space != -1 {
			word = line[:space]
			for space < len(line)-1 && line[space+1] == ' ' {
				space++
			}
			line = line[space+1:]
		} else {
			word = line[:]
			line = line[len(line):]
		}
	}

	// Handle empty lines.
	if line == nil || len(line) == 0 {
		return
	}

	// Find the first word.
	nextword()

	// If it begins with a ':', it's an origin.
	// Note it aside and step on to the next word.
	if len(word) > 0 && word[0] == ':' {
		origin = word[1:]

		nextword()
	}

	// The word we have now is the command.
	// Look it up. If it doesn't exist, return early.
	cmdName := string(word)
	command = d.Lookup(cmdName)
	if command == nil {
		err = newParseError(CmdNotFound, cmdName)
		return
	}

	// Check the command is for our registration status.
	if command.Unregged == 0 && !regged {
		command = nil
		err = newParseError(CmdForRegistered, cmdName)
		return
	}
	if command.Unregged == 2 && regged {
		command = nil
		err = newParseError(CmdForUnregistered, cmdName)
		return
	}

	// If the command doesn't support parameters, skip them all.
	if command.Maxargs == 0 {
		return
	}

	// Everything else is a sequence of parameters.
	for len(params) < 50 {

		// If the line is empty, break.
		if len(line) == 0 {
			break
		}

		// If it begins with a colon, the rest of the line after that
		// point is the final parameter.
		if line[0] == ':' {
			param_array[len(params)] = line[1:]
			params = params[:len(params)+1]
			break
		}

		// If we've hit the limit for parameters, the whole rest of the
		// line is one large final parameter.
		if len(params) == command.Maxargs-1 || len(params) == 49 {
			param_array[len(params)] = line
			params = params[:len(params)+1]
			break
		}

		// Otherwise, this parameter runs up to the next space.
		nextword()
		param_array[len(params)] = word
		params = params[:len(params)+1]
	}

	// If we don't have enough parameters, treat it as a failed dispatch.
	if len(params) < command.Minargs {
		err = newParseError(CmdTooFewParams, cmdName)
		command = nil
		params = param_array[:0]
	}

	return
}


// Represents an error parsing a line.
type ParseError struct {
	Num     int
	CmdName string
}

// Represents specific parser errors.
const (
	CmdNotFound = iota
	CmdForRegistered
	CmdForUnregistered
	CmdTooFewParams
)

// String returns an error message for the parse error.
// It also makes ParseError meet the os.Error interface.
func (err *ParseError) String() string {
	switch err.Num {
	case CmdNotFound:
		return "Unknown command."
	case CmdForRegistered:
		return "You have not registered."
	case CmdForUnregistered:
		return "You may not reregister."
	case CmdTooFewParams:
		return "Not enough parameters."
	}

	// We don't really know what happened, so bullshit them.
	return "Invalid command."
}

func newParseError(num int, cmdName string) (err *ParseError) {
	err = new(ParseError)
	err.Num = num
	err.CmdName = cmdName
	return
}
