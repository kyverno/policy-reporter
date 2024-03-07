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

func Merge[T comparable, R any](first, second map[T]R) map[T]R {
	merged := make(map[T]R, len(first)+len(second))

	for k, v := range first {
		merged[k] = v
	}
	for k, v := range second {
		merged[k] = v
	}

	return merged
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

func ConvertMap(m map[string]any) map[string]string {
	n := make(map[string]string, len(m))
	for k, v := range m {
		if l, ok := v.(string); ok {
			n[k] = l
		}
	}

	return n
}
