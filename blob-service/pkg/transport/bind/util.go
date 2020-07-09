package bind

import "strings"

func getServiceFromPath(url string) string {
	x := strings.Split(url, "/")
	if len(x) >= 2 {
		return x[len(x)-2]
	}

	return ""
}
