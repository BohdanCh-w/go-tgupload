// nolint: ireturn
package utils

func DefaultIfNil[T any](val *T, def T) T {
	if val == nil {
		return def
	}

	return *val
}
