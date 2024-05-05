package helper

import (
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

func ConvertMap(m map[string]any) map[string]string {
	n := make(map[string]string, len(m))
	for k, v := range m {
		if l, ok := v.(string); ok {
			n[k] = l
		}
	}

	return n
}

func Defaults(s, f string) string {
	if s != "" {
		return s
	}

	return f
}

func ToPointer[T any](s T) *T {
	return &s
}

func Filter[T any](s []T, keep func(T) bool) []T {
	d := make([]T, 0, len(s))
	for _, n := range s {
		if keep(n) {
			d = append(d, n)
		}
	}
	return d
}
