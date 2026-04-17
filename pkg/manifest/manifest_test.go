package manifest

import "testing"

func TestUsesAssets(t *testing.T) {
	m := &Manifest{}
	if m.UsesAssets() {
		t.Fatal("empty manifest should not report assets")
	}
	m.Assets = map[string]Asset{"linux/amd64": {URL: "https://example/x"}}
	if !m.UsesAssets() {
		t.Fatal("manifest with assets should report assets")
	}
}

func TestAssetFor(t *testing.T) {
	m := &Manifest{
		Assets: map[string]Asset{
			"linux/amd64":  {URL: "https://example/linux"},
			"darwin/arm64": {URL: "https://example/darwin"},
		},
	}
	if _, ok := m.AssetFor("linux/amd64"); !ok {
		t.Fatal("expected asset for linux/amd64")
	}
	if _, ok := m.AssetFor("windows/amd64"); ok {
		t.Fatal("did not expect asset for windows/amd64")
	}
}

func TestResolvedBinariesDefault(t *testing.T) {
	m := &Manifest{Name: "ffuf"}
	bins := m.ResolvedBinaries()
	if len(bins) != 1 || bins[0] != "ffuf" {
		t.Fatalf("unexpected default binaries: %v", bins)
	}
}

func TestResolvedBinariesExplicit(t *testing.T) {
	m := &Manifest{Name: "foo", Binaries: []string{"foo", "foo-helper"}}
	bins := m.ResolvedBinaries()
	if len(bins) != 2 || bins[1] != "foo-helper" {
		t.Fatalf("unexpected explicit binaries: %v", bins)
	}
}

func TestToYAMLRoundTrip(t *testing.T) {
	m := &Manifest{
		Name:    "ffuf",
		Version: "2.1.0",
		Assets: map[string]Asset{
			"linux/amd64": {URL: "https://example", Checksum: "sha256:abc", Type: "tar.gz"},
		},
	}
	data, err := ToYAML(m)
	if err != nil {
		t.Fatalf("ToYAML failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("ToYAML produced empty output")
	}
}
