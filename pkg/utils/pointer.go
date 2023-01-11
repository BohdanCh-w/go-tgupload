package utils

func PtrOrNil[T comparable](val T) *T {
	var zero T

	if val == zero {
		return nil
	}

	return &val
}
