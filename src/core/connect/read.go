package connect

import proto "goprotobuf.googlecode.com/hg/proto"

import "oddcomm/src/core/connect/mmn"

// Reads incoming lines from the given connection, and sends them
// on the given channel. Closes the channel when the connection is closed.
func (conn *Conn) ReadLines(ch chan<- *mmn.Line) {
	fullReadBuffer := make([]byte, 0, 10240)
	readBuffer := fullReadBuffer
	readLengthMode := true
	var readLength int = 0
	for {
		// Read from the connection.
		remaining := readBuffer[len(readBuffer):cap(readBuffer)]
		n, err := conn.conn.Read(remaining)
		readBuffer = readBuffer[:len(readBuffer)+n]
		if err != nil {
			conn.Close()
			close(ch)
			return
		}

		// Handle reading message length.
		if readLengthMode {
			length, start := proto.DecodeVarint(readBuffer)
			if start != 0 {

				// Check for overlength lines.
				if length > 10240 {
					conn.Close()
					close(ch)
					return
				}

				// Copy down the buffer, start reading msg.
				copy(readBuffer, readBuffer[start:])
				readBuffer = readBuffer[:len(readBuffer)-start]
				readLength = int(length)
				readLengthMode = false

			} else {
				// Not done reading the length, read more.
				continue
			}
		}

		// If we've not read the full line, read more.
		if len(readBuffer) < int(readLength) {
			continue
		}

		// Parse the line.
		lineBuffer := readBuffer[:readLength]
		line := new(mmn.Line)
		err = proto.Unmarshal(lineBuffer, line)
		if err != nil {
			conn.Close()
			close(ch)
			return
		}

		// Send the line to be handled.
		ch <- line

		// Copy down the remainder of the buffer,
		// and reduce length to that.
		copy(readBuffer, readBuffer[readLength:])
		readBuffer = readBuffer[:len(readBuffer) - int(readLength)]

		// Start reading the length of the next line.
		readLengthMode = true
	}
}
