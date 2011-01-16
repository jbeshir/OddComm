package perm

import "os"


// Run a permission hook list.
// Returns whether the action is to be permitted or denied. def specifies the
// default value.
func runPermHooks(l []interface{}, f func(interface{}) (int, os.Error), def bool) (perm int, err os.Error) {
	var absperm int = 1
	if def {
		perm = 1
	} else {
		perm = -1
	}

	for _, h := range l {
		result, thisErr := f(h)

		// A result of -1e9 or below indicates that the operation will
		// break the server if it is permitted. This is an automatic,
		// not-overridable "No.".
		if result <= -1e9 {
			return -1e9, thisErr
		}

		// Take the largest result in magnitude. If they are equal in
		// magnitude, take the positive.
		absresult := result
		if result < 0 {
			absresult = -result
		}
		if absresult > absperm {
			perm = result
			absperm = absresult
			err = thisErr
		} else if absresult == absperm && result > 0 {
			perm = result
			absperm = absresult
			err = thisErr
		}
	}

	if perm < 0 && err == nil {
		err = os.NewError("Permission denied.")
	}

	return
}

// Run a slice of permission hook lists, and combine results.
func runPermHooksSlice(lists [][]interface{}, f func(interface{}) (int, os.Error), def bool) (perm int, err os.Error) {
	if def {
		perm = 1
	} else {
		perm = -1
	}

	var absPerm int
	for _, l := range lists {
		thisPerm, thisErr := runPermHooks(l, f, def)
		if thisPerm <= -1e9 {
			return -1e9, thisErr
		}

		absThisPerm := thisPerm
		if thisPerm < 0 {
			absThisPerm = -thisPerm
		}

		if absThisPerm > absPerm {
			perm = thisPerm
			absPerm = absThisPerm
			err = thisErr
		} else if absThisPerm == absPerm && thisPerm > 0 {
			perm = thisPerm
			absPerm = absThisPerm
			err = thisErr
		}
	}

	if perm < 0 && err == nil {
		err = os.NewError("Permission denied.")
	}

	return
}
