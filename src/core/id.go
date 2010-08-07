package core


var currentUID int64


func incrementUID() {
	currentUID++
	if currentUID > 36*36*36*36*36*36-1 {
		currentUID = 0
	}
}

func getUIDString() string {
	var buf [6]byte
	id := currentUID
	for i := 5; i >= 0; i-- {
		char := uint8(id % 36)
		if char < 10 {
			buf[i] = 48 + char
		} else {
			buf[i] = 55 + char
		}
		id = id/36
	}
	return "1AA" + string(buf[0:])
}
