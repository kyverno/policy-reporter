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

func ToList[T any, R comparable](mapping map[R]T) []T {
	list := make([]T, 0, len(mapping))
	for _, i := range mapping {
		list = append(list, i)
	}

	return list
}
