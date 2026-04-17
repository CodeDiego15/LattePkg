package archive

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func writeTarGz(t *testing.T, dst string, entries map[string]string) {
	t.Helper()
	f, err := os.Create(dst)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	for name, body := range entries {
		hdr := &tar.Header{
			Name: name,
			Mode: 0o755,
			Size: int64(len(body)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("header: %v", err)
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("tar close: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("gz close: %v", err)
	}
}

func TestExtractTarGz(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "pkg.tar.gz")
	writeTarGz(t, archivePath, map[string]string{"bin/ffuf": "binary"})

	dest := filepath.Join(dir, "out")
	if err := Extract(archivePath, dest, "tar.gz", 0); err != nil {
		t.Fatalf("Extract: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(dest, "bin", "ffuf"))
	if err != nil {
		t.Fatalf("read extracted: %v", err)
	}
	if string(b) != "binary" {
		t.Fatalf("unexpected content: %s", b)
	}
}

func TestExtractRejectsPathTraversal(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "bad.tar.gz")
	writeTarGz(t, archivePath, map[string]string{"../../evil.sh": "pwn"})

	dest := filepath.Join(dir, "out")
	if err := Extract(archivePath, dest, "tar.gz", 0); err == nil {
		t.Fatal("expected path traversal to be rejected")
	}
}

func TestInferType(t *testing.T) {
	cases := map[string]string{
		"foo.tar.gz":  "tar.gz",
		"foo.tgz":     "tar.gz",
		"foo.zip":     "zip",
		"foo.tar":     "tar",
		"foo.tar.bz2": "tar.bz2",
		"foo":         "raw",
	}
	for in, want := range cases {
		if got := InferType(in); got != want {
			t.Errorf("InferType(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestExtractZip(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "pkg.zip")

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	w, err := zw.Create("tool")
	if err != nil {
		t.Fatalf("zip create: %v", err)
	}
	if _, err := w.Write([]byte("hi")); err != nil {
		t.Fatalf("zip write: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	if err := os.WriteFile(archivePath, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write zip: %v", err)
	}

	dest := filepath.Join(dir, "out")
	if err := Extract(archivePath, dest, "zip", 0); err != nil {
		t.Fatalf("Extract: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(dest, "tool"))
	if err != nil {
		t.Fatalf("read extracted: %v", err)
	}
	if string(b) != "hi" {
		t.Fatalf("unexpected: %s", b)
	}
}
