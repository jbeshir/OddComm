package client


type hook struct {
	next *hook
	h    func() string
}


var supportLine string
var supportHooks *hook


func init() {
	// Add core support items.
	AddSupport("CASEMAPPING=ascii")
	AddSupport("CLIENTVER=3.0")
	AddSupport("CHARSET=UTF-8")
	AddSupport("EXCEPTS=e")
	AddSupport("FNC")
	AddSupport("INVEX=I")
	AddSupport("NAMESX")
	AddSupport("NETWORK=Testnet")

	// Add "fake" support items. These are not really implemented, as we
	// can pause within execution of a command to ratelimit between targets
	// and prefer this over limiting targets.
	AddSupport("MAXTARGETS=100")
	AddSupport("MODES=1000")

	// Add core support hooks.
	AddSupportHook(func() string {
		prefixes, modes := ChanModes.AllPrefixes()
		return " PREFIX=(" + modes + ")" + prefixes
	})
	AddSupportHook(func() string {
		return " CHANMODES=" + ChanModes.AllList() + ",," +
			ChanModes.AllParametered() + "," + ChanModes.AllSimple()
	})
}


// Adds the given token to the server's ISUPPORT string.
// May only be used during init.
func AddSupport(value string) {
	if supportLine != "" {
		supportLine += " " + value
	} else {
		supportLine = value
	}
}

// Adds a ISUPPORT hook. Called after init is complete to return a string which
// will be appended to the ISUPPORT string. If the returned value is non-empty,
// it must start with a space, to separate it from previous values.
// May only be used during init.
func AddSupportHook(h func() string) {
	hook := new(hook)
	hook.h = h
	hook.next = supportHooks
	supportHooks = hook
}
