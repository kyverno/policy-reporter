package helper

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strings"
)

var (
	Caser = cases.Title(language.English, cases.NoLower)
)

func Contains(source string, sources []string) bool {
	for _, s := range sources {
		if strings.EqualFold(s, source) {
			return true
		}
	}

	return false
}
