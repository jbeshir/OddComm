package irc

import "bytes"

func Parse(d CommandDispatcher, line []byte) (origin []byte, command *Command, params [][]byte) {

	// We can handle up to 20 parameters. This is plenty.
	var param_array [20][]byte
	params = param_array[0:0]
	var word []byte

	// Define function for moving to the next word.
	var nextword = func() {
		space := bytes.IndexByte(line, ' ')
		if space != -1 {
			word = line[0:space]
			for space < len(line)-1 && line[space+1] == ' ' {
				space++
			}
			line = line[space+1:]
		} else {
			word = line[0:]
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
	if (len(word) > 0 && word[0] == ':') {
		origin = word[1:]

		nextword()
	}

	// The word we have now is the command.
	// Look it up. If it doesn't exist, return early.
	command = d.Lookup(string(word))
	if command == nil {
		return
	}

	// If the command doesn't support parameters, skip them all.
	if command.Maxargs == 0 {
		return
	}

	// Everything else is a sequence of parameters.
	for len(params) < 20 {

		// If the line is empty, break.
		if (len(line) == 0) {
			break
		}

		// If it begins with a colon, the rest of the line after that
		// point is the final parameter.
		if (line[0] == ':') {
			param_array[len(params)] = line[1:]
			params = params[0:len(params)+1]
			break
		}

		// If we've hit the limit for parameters, the whole rest of the
		// line is one large final parameter.
		if len(params) == command.Maxargs - 1 || len(params) == 19 {
			param_array[len(params)] = line
			params = params[0:len(params)+1]
			break
		}

		// Otherwise, this parameter runs up to the next space.
		nextword()
		param_array[len(params)] = word
		params = params[0:len(params)+1]
	}

	// If we don't have enough parameters, treat it as a failed dispatch.
	if len(params) < command.Minargs {
		command = nil
		params = param_array[0:0]
	}

	return
}
