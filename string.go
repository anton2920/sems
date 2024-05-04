package main

//go:noescape
//go:nosplit
func FindChar(s string, c byte) int

//go:noescape
//go:nosplit
func FindSubstring(a, b string) int

/* TODO(anton2920): rewrite using (DF=1 and REP SCASB) or SIMD. */
func FindCharReverse(s string, c byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == c {
			return i
		}
	}

	return -1
}
