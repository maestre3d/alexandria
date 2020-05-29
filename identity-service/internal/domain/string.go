package domain

import "strings"

// CapitalizeString capitalize the string's first character
// i.e. ethan -> Ethan
func CapitalizeString(s string) string {
	if len(s) == 1 {
		return strings.ToUpper(s)
	}
	return strings.ToUpper(s[0:1]) + s[1:]
}
