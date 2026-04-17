// Package integrity implements checksum verification for downloaded assets.
//
// Only SHA-256 is supported today. Checksums are expressed as
// "<algorithm>:<hex>" strings inside manifests.
package integrity

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

// ErrMismatch is returned when a checksum does not match the expected value.
var ErrMismatch = fmt.Errorf("checksum mismatch")

// Parse splits an "<algo>:<hex>" checksum into its components. It returns
// an error if the algorithm is not recognised or the hex part is empty.
func Parse(s string) (algo, sum string, err error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid checksum format %q: expected <algo>:<hex>", s)
	}
	algo = strings.ToLower(strings.TrimSpace(parts[0]))
	sum = strings.ToLower(strings.TrimSpace(parts[1]))
	if sum == "" {
		return "", "", fmt.Errorf("invalid checksum %q: empty digest", s)
	}
	switch algo {
	case "sha256":
		// supported
	default:
		return "", "", fmt.Errorf("unsupported checksum algorithm %q", algo)
	}
	return algo, sum, nil
}

// VerifyFile checks that the contents of path match expected. expected must
// be of the form "<algo>:<hex>".
func VerifyFile(path, expected string) error {
	algo, want, err := Parse(expected)
	if err != nil {
		return err
	}
	got, err := hashFile(path, algo)
	if err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf("%w: expected %s got %s:%s", ErrMismatch, expected, algo, got)
	}
	return nil
}

// SumFile returns the "<algo>:<hex>" digest of path.
func SumFile(path, algo string) (string, error) {
	digest, err := hashFile(path, algo)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s", algo, digest), nil
}

// HashBytes returns the "<algo>:<hex>" digest of data. It panics if algo
// is unsupported; callers that want to recover from an unsupported
// algorithm should use Parse first.
func HashBytes(algo string, data []byte) string {
	switch algo {
	case "sha256":
		sum := sha256.Sum256(data)
		return fmt.Sprintf("%s:%s", algo, hex.EncodeToString(sum[:]))
	default:
		panic(fmt.Sprintf("integrity: unsupported algorithm %q", algo))
	}
}

func hashFile(path, algo string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	switch algo {
	case "sha256":
		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			return "", fmt.Errorf("hash %s: %w", path, err)
		}
		return hex.EncodeToString(h.Sum(nil)), nil
	default:
		return "", fmt.Errorf("unsupported algorithm %q", algo)
	}
}
