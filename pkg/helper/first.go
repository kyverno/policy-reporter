package helper

func First[T any](list []*T) *T {
	if len(list) == 0 {
		return nil
	}

	return list[0]
}
