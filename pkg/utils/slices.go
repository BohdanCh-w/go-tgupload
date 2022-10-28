package utils

func EnumeratedMap[T comparable](slice []T) map[T]int {
	ret := make(map[T]int, len(slice))

	for i, val := range slice {
		ret[val] = i
	}

	return ret
}
