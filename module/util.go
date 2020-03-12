package module

import "strings"

func filterNulls(s string) string {
	return strings.Replace(s, "\x00", "", -1)
}
