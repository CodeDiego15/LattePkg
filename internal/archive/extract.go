// Package archive extracts tar / tar.gz / tar.bz2 / zip archives.
// Raw binaries ("raw") are copied verbatim into the destination directory
// with the same base name as the source file.
package archive

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/DiegoDev2/Fleet/pkg/manifest"
)

// InferType guesses the archive type from a URL or filename when the
// manifest does not set one explicitly. "raw" is returned for unknown
// extensions on the assumption that the file is a standalone binary.
func InferType(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, ".tar.gz"), strings.HasSuffix(lower, ".tgz"):
		return "tar.gz"
	case strings.HasSuffix(lower, ".tar.bz2"), strings.HasSuffix(lower, ".tbz2"):
		return "tar.bz2"
	case strings.HasSuffix(lower, ".tar"):
		return "tar"
	case strings.HasSuffix(lower, ".zip"):
		return "zip"
	default:
		return "raw"
	}
}

// Extract unpacks src into destDir. kind should be one of the constants
// accepted by InferType.
func Extract(src, destDir, kind string, strip int) error {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("create dest dir: %w", err)
	}
	switch kind {
	case "tar.gz", "tgz":
		return untar(src, destDir, "gzip", strip)
	case "tar.bz2", "tbz2":
		return untar(src, destDir, "bzip2", strip)
	case "tar":
		return untar(src, destDir, "", strip)
	case "zip":
		return unzip(src, destDir, strip)
	case "raw", "":
		return copyRaw(src, destDir)
	default:
		return fmt.Errorf("unsupported archive type %q", kind)
	}
}

// ExtractAsset is a convenience wrapper that reads the archive type and
// strip setting from the manifest asset.
func ExtractAsset(src, destDir string, asset manifest.Asset) error {
	kind := asset.Type
	if kind == "" {
		kind = InferType(src)
	}
	return Extract(src, destDir, kind, asset.StripComponents)
}

func untar(src, destDir, compression string, strip int) error {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer f.Close()

	var r io.Reader = f
	switch compression {
	case "gzip":
		gr, err := gzip.NewReader(f)
		if err != nil {
			return fmt.Errorf("gzip: %w", err)
		}
		defer gr.Close()
		r = gr
	case "bzip2":
		r = bzip2.NewReader(f)
	case "":
		// plain tar
	default:
		return fmt.Errorf("unsupported compression %q", compression)
	}

	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar: %w", err)
		}

		name := stripComponents(hdr.Name, strip)
		if name == "" {
			continue
		}
		target, err := securePath(destDir, name)
		if err != nil {
			return err
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(hdr.Mode)|0o700); err != nil {
				return fmt.Errorf("mkdir %s: %w", target, err)
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return fmt.Errorf("mkdir %s: %w", filepath.Dir(target), err)
			}
			if err := writeFile(target, tr, os.FileMode(hdr.Mode)|0o600); err != nil {
				return err
			}
		case tar.TypeSymlink:
			// Skip symlinks for safety in this MVP.
			continue
		default:
			// Skip other entry types.
			continue
		}
	}
	return nil
}

func unzip(src, destDir string, strip int) error {
	zr, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer zr.Close()

	for _, f := range zr.File {
		name := stripComponents(f.Name, strip)
		if name == "" {
			continue
		}
		target, err := securePath(destDir, name)
		if err != nil {
			return err
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return fmt.Errorf("mkdir %s: %w", target, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Dir(target), err)
		}
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("open zip entry %s: %w", f.Name, err)
		}
		err = writeFile(target, rc, f.Mode()|0o600)
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func copyRaw(src, destDir string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open raw: %w", err)
	}
	defer in.Close()

	target := filepath.Join(destDir, filepath.Base(src))
	return writeFile(target, in, 0o755)
}

func writeFile(path string, r io.Reader, mode os.FileMode) error {
	out, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	if _, err := io.Copy(out, r); err != nil {
		out.Close()
		return fmt.Errorf("write %s: %w", path, err)
	}
	return out.Close()
}

func stripComponents(name string, strip int) string {
	if strip <= 0 {
		return filepath.Clean(name)
	}
	parts := strings.Split(filepath.ToSlash(filepath.Clean(name)), "/")
	if len(parts) <= strip {
		return ""
	}
	return filepath.Join(parts[strip:]...)
}

// securePath joins base and name, rejecting paths that escape base.
// The check relies on filepath.Rel, which is the canonical way to decide
// whether a cleaned path lies inside base on any platform.
func securePath(base, name string) (string, error) {
	if filepath.IsAbs(name) {
		return "", fmt.Errorf("refusing to extract %q: absolute path", name)
	}
	cleaned := filepath.Clean(name)
	target := filepath.Join(base, cleaned)
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return "", fmt.Errorf("invalid path %q: %w", name, err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("refusing to extract %q: path traversal", name)
	}
	return target, nil
}
