package perm

// A permissions hook. Contains a function to run.
type hook struct {
	next *hook
	f interface{}
}

// Add a permissions hook to a hook list.
func hookAdd(l **hook, f interface{}) {
	h := new(hook)
	h.f = f
	h.next = *l
	*l = h
}

// Run a permission hook list.
// Returns whether the action is to be permitted or denied. def specifies the
// default value.
func (l *hook) run(f func(interface{}) int, def bool) (perm int) {
	var absperm int = 1
	if def {
		perm = 1
	} else {
		perm = -1
	}
	
	for h := l; h != nil; h = h.next {
		result := f(h.f)

		// A result of -1e9 or below indicates that the operation will
		// break the server if it is permitted. This is an automatic,
		// not-overridable "No.".
		if result <= -1e9 {
			return -1e9
		}

		// Take the largest result in magnitude. If they are equal in
		// magnitude, take the positive.
		absresult := result
		if result < 0 {
			absresult = -result
		}
		if absresult > absperm {
			perm = result
		} else if absresult == absperm && result < 0 {
			perm = result
		}
	}

	return
}

// Run a slice of permission hook lists, and combine results.
func runPermHookLists(lists []*hook, f func(interface{}) int,
                      def bool) (perm int) {
	var absPerm int
	for i := range lists {
		thisPerm := lists[i].run(f, def)
		if thisPerm <= -1e9 {
			return -1e9
		}

		absThisPerm := thisPerm
		if thisPerm < 0 {
			absThisPerm = -thisPerm
		}

		if absThisPerm > absPerm {
			perm = thisPerm
			absPerm = absThisPerm
		} else if absThisPerm == absPerm && thisPerm > 0 {
			perm = thisPerm
			absPerm = absThisPerm
		}
	}


	return
}
