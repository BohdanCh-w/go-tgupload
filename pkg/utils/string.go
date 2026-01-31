package utils

import "strings"

func MaskString(s string) string {
	const (
		maxShown  = 6
		partShown = 0.25
	)

	if s == "" {
		return s
	}

	lenShown := int(float64(len(s)) * partShown)
	lenMasked := len(s) - lenShown

	return strings.Repeat("*", lenMasked) + s[lenMasked:]
}
