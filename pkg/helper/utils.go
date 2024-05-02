package helper

import "strings"

func Contains(source string, sources []string) bool {
	for _, s := range sources {
		if strings.EqualFold(s, source) {
			return true
		}
	}

	return false
}

func Defaults(s, f string) string {
	if s != "" {
		return s
	}

	return f
}
