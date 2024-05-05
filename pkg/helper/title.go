package helper

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var caser = cases.Title(language.English, cases.NoLower)

func Title(s string) string {
	return caser.String(s)
}
