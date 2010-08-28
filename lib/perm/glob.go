package perm

import "regexp"
import "strings"


// GMatch matches the given value against the given glob mask.
// The implementation is a placeholder, pending a proper one.
func GMatch(value, mask string) bool {
	mask = regexp.QuoteMeta(mask)
	mask = strings.Replace(mask, `\*`, ".*", -1)
	mask = strings.Replace(mask, `\?`, ".", -1)
	mask = strings.Replace(mask, `\[`, "[", -1)
	mask = strings.Replace(mask, `\]`, "]", -1)
	if r, _ := regexp.Compile(mask); r != nil {
		return r.MatchString(value)
	}
	return false
}
