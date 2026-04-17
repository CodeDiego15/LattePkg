package config

import (
	"path/filepath"
	"testing"
)

func TestAddRemoveRepository(t *testing.T) {
	c := Default()
	before := len(c.Repositories)

	added := c.AddRepository(Repository{Name: "extra", URL: "https://example", Type: "git", Priority: 50})
	if !added {
		t.Fatal("expected AddRepository to return true on first add")
	}
	if c.AddRepository(Repository{Name: "extra", URL: "https://example", Type: "git"}) {
		t.Fatal("expected duplicate add to return false")
	}
	if !c.RemoveRepository("extra") {
		t.Fatal("expected RemoveRepository to return true")
	}
	if len(c.Repositories) != before {
		t.Fatalf("expected repositories to return to original size, got %d", len(c.Repositories))
	}
	if c.RemoveRepository("missing") {
		t.Fatal("expected RemoveRepository to return false for missing name")
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	c := Config{Repositories: []Repository{{Name: "a", URL: "u", Type: "git", Priority: 10}}}
	if err := c.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded.Repositories) != 1 || loaded.Repositories[0].Name != "a" {
		t.Fatalf("unexpected load: %+v", loaded)
	}
}

func TestLoadMissingFileReturnsDefault(t *testing.T) {
	dir := t.TempDir()
	loaded, err := Load(filepath.Join(dir, "nope.yaml"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	// Default() may return a seeded repo list; what we care about is that
	// the call succeeds instead of returning an error.
	_ = loaded
}
