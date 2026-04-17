package state

import (
	"path/filepath"
	"testing"
	"time"
)

func TestStoreRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")

	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if got := s.List(); len(got) != 0 {
		t.Fatalf("expected empty store, got %d", len(got))
	}

	if err := s.Put(Installed{
		Name:        "ffuf",
		Version:     "2.1.0",
		Platform:    "linux/amd64",
		InstalledAt: time.Now().UTC(),
		PackageDir:  "/pkg/ffuf/2.1.0",
		Binaries:    []string{"/bin/ffuf"},
	}); err != nil {
		t.Fatalf("Put: %v", err)
	}

	s2, err := Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	got, ok := s2.Get("ffuf")
	if !ok {
		t.Fatal("expected ffuf to be persisted")
	}
	if got.Version != "2.1.0" {
		t.Fatalf("unexpected version: %s", got.Version)
	}

	if err := s2.Delete("ffuf"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, ok := s2.Get("ffuf"); ok {
		t.Fatal("expected ffuf to be deleted")
	}
}
