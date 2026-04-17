package integrity

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	algo, sum, err := Parse("sha256:abcd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if algo != "sha256" || sum != "abcd" {
		t.Fatalf("unexpected parse output: %s / %s", algo, sum)
	}
	if _, _, err := Parse("md5:abcd"); err == nil {
		t.Fatal("expected error for md5")
	}
	if _, _, err := Parse("bogus"); err == nil {
		t.Fatal("expected error for missing colon")
	}
}

func TestHashBytesAndVerify(t *testing.T) {
	data := []byte("hello armada")
	sum := HashBytes("sha256", data)

	dir := t.TempDir()
	path := filepath.Join(dir, "hello.txt")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write temp: %v", err)
	}

	if err := VerifyFile(path, sum); err != nil {
		t.Fatalf("VerifyFile(valid): %v", err)
	}
	if err := VerifyFile(path, "sha256:deadbeef"); err == nil {
		t.Fatal("expected mismatch error")
	}
}
