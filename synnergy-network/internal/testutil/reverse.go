package testutil

// Reverse returns its input string with bytes in reverse order.
// It works on raw bytes so it is an involution even for invalid UTF-8.
func Reverse(s string) string {
	b := []byte(s)
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return string(b)
}
