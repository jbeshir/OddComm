package ts6

import "os"

import "oddcomm/src/core"


var all allSender
type allSender struct {
}

func (a allSender) Write(b []byte) (n int, err os.Error) {
	core.IterateSID(func(sid string, value interface{}) {
		s, ok := value.(*server)
		if !ok {
			return
		}

		if &(s.local.server) != s {
			return
		}

		if s.local.bursted != true {
			return
		}

		s.local.Write(b)
	})

	return len(b), nil
}
