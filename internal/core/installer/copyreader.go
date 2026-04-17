package installer

import "io"

// copyReader wraps io.Copy so we can stub it in tests if needed.
func copyReader(dst io.Writer, src io.Reader) (int64, error) {
	return io.Copy(dst, src)
}
