package helper

import (
	"encoding/json"
	"strings"
)

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

func Map[T any, R any](source []T, cb func(T) R) []R {
	list := make([]R, 0, len(source))
	for _, i := range source {
		list = append(list, cb(i))
	}

	return list
}

func ConvertJSONToMap(s string) map[string]string {
	m := make(map[string]string)
	if s == "" {
		return m
	}

	_ = json.Unmarshal([]byte(s), &m)

	return m
}
