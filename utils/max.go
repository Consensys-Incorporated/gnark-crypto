package utils

// Max returns the maximum of two integers.
//
// Deprecated: use the max builtin. Kept until the next major release to avoid
// breaking vendored code.
//
//go:fix inline
func Max(a, b int) int {
	return max(a, b)
}
