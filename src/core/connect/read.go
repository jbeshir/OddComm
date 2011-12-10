package connect

import proto "goprotobuf.googlecode.com/hg/proto"

import "oddcomm/src/core/connect/mmn"

// Reads incoming lines from the given connection, and sends them
// on the given channel. Closes the channel when the connection is closed.
func (conn *Conn) ReadLines(ch chan<- *mmn.Line) {
	readBuffer := make([]byte, 0, 10240)
	readLengthMode := true
	var readLength uint64 = 0
	for {
		// Read from the connection.
		n, err := conn.conn.Read(readBuffer[len(readBuffer):])
		readBuffer = readBuffer[:len(readBuffer)+n]
		if err != nil {
			conn.Close()
			close(ch)
			return
		}

		// If we don't have a length for the next message, read length.
		if readLengthMode {
			var start int
			readLength, start = proto.DecodeVarint(readBuffer)
			if (start != 0) {
				copy(readBuffer, readBuffer[start:])
				readBuffer = readBuffer[:len(readBuffer)-start]
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

		// Parse the line and call the handler.
		lineBuffer := readBuffer[:readLength]
		line := new(mmn.Line)
		err = proto.Unmarshal(lineBuffer, line)
		if err != nil {
			conn.Close()
			close(ch)
			return
		}

		ch <- line

		// Go back to reading line length.
		readLengthMode = true
	}
}
