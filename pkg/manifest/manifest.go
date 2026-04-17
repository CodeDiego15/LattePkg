// Package manifest defines the public Armada manifest schema.
//
// Manifests are YAML files that describe how to install a tool on one or
// more platforms. The canonical format uses the `assets` map keyed by
// "<os>/<arch>" and a list of `binaries` to symlink.
//
// A legacy nested format (`platforms[os].architecture[arch]`) is still
// parsed for backwards compatibility with manifests written for earlier
// versions of the project.
package manifest

import "gopkg.in/yaml.v3"

// Manifest describes an installable tool.
type Manifest struct {
	Name        string   `yaml:"name" json:"name"`
	Version     string   `yaml:"version" json:"version"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
	Homepage    string   `yaml:"homepage,omitempty" json:"homepage,omitempty"`
	License     string   `yaml:"license,omitempty" json:"license,omitempty"`
	Categories  []string `yaml:"categories,omitempty" json:"categories,omitempty"`

	// Assets is the preferred installation source. Keys use the form
	// "<os>/<arch>" (e.g. "linux/amd64", "darwin/arm64").
	Assets map[string]Asset `yaml:"assets,omitempty" json:"assets,omitempty"`

	// Binaries is the list of executable names that Armada will symlink
	// into its bin directory after a successful install. Each entry is
	// resolved against the extracted asset directory. If Binaries is
	// empty, Armada defaults to symlinking a single binary named like
	// the manifest (Name).
	Binaries []string `yaml:"binaries,omitempty" json:"binaries,omitempty"`

	// Legacy fields kept for backwards compatibility with older manifests.
	Platforms    map[string]Platform `yaml:"platforms,omitempty" json:"platforms,omitempty"`
	Dependencies []Dependency        `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
	PostInstall  []string            `yaml:"post_install,omitempty" json:"post_install,omitempty"`
}

// Asset describes a single downloadable artifact for one os/arch.
type Asset struct {
	// URL of the archive or raw binary to download.
	URL string `yaml:"url" json:"url"`

	// Checksum in the form "<algo>:<hex>". Only sha256 is supported today.
	Checksum string `yaml:"checksum" json:"checksum"`

	// Type overrides the archive type. Valid values: "tar.gz", "tgz",
	// "tar.xz", "tar.bz2", "zip", "raw". When empty, Armada infers the
	// type from the URL extension.
	Type string `yaml:"type,omitempty" json:"type,omitempty"`

	// StripComponents mirrors `tar --strip-components` for tar archives.
	StripComponents int `yaml:"strip_components,omitempty" json:"strip_components,omitempty"`
}

// Platform is part of the legacy manifest format.
type Platform struct {
	Architecture    map[string]Architecture   `yaml:"architecture,omitempty" json:"architecture,omitempty"`
	PackageManagers map[string]PackageManager `yaml:"package_managers,omitempty" json:"package_managers,omitempty"`
	Fallback        *Fallback                 `yaml:"fallback,omitempty" json:"fallback,omitempty"`
}

// Architecture is part of the legacy manifest format.
type Architecture struct {
	URL          string `yaml:"url" json:"url"`
	Type         string `yaml:"type" json:"type"`
	Checksum     string `yaml:"checksum" json:"checksum"`
	InstallSteps []Step `yaml:"install_steps" json:"install_steps"`
}

// Step is part of the legacy manifest format.
type Step struct {
	Mount   string `yaml:"mount,omitempty" json:"mount,omitempty"`
	Copy    string `yaml:"copy,omitempty" json:"copy,omitempty"`
	To      string `yaml:"to,omitempty" json:"to,omitempty"`
	Execute string `yaml:"execute,omitempty" json:"execute,omitempty"`
	Unmount string `yaml:"unmount,omitempty" json:"unmount,omitempty"`
}

// PackageManager is part of the legacy manifest format.
type PackageManager struct {
	Package string `yaml:"package" json:"package"`
}

// Fallback is part of the legacy manifest format.
type Fallback struct {
	URL          string               `yaml:"url" json:"url"`
	Type         string               `yaml:"type" json:"type"`
	Checksum     string               `yaml:"checksum" json:"checksum"`
	Dependencies FallbackDependencies `yaml:"dependencies" json:"dependencies"`
	BuildSteps   []string             `yaml:"build_steps" json:"build_steps"`
}

// FallbackDependencies is part of the legacy manifest format.
type FallbackDependencies struct {
	Build   []string `yaml:"build" json:"build"`
	Runtime []string `yaml:"runtime" json:"runtime"`
}

// Dependency describes a required sibling tool.
type Dependency struct {
	Name    string `yaml:"name" json:"name"`
	Version string `yaml:"version" json:"version"`
}

// UsesAssets reports whether the manifest uses the modern `assets` format.
func (m *Manifest) UsesAssets() bool {
	return len(m.Assets) > 0
}

// AssetFor returns the asset matching the given "os/arch" key, if any.
func (m *Manifest) AssetFor(osArch string) (Asset, bool) {
	a, ok := m.Assets[osArch]
	return a, ok
}

// ResolvedBinaries returns the list of binary names that should be
// symlinked into the Armada bin directory after install. If the manifest
// does not declare any binaries, the manifest Name is used as the single
// default binary.
func (m *Manifest) ResolvedBinaries() []string {
	if len(m.Binaries) > 0 {
		return m.Binaries
	}
	return []string{m.Name}
}

// ToYAML serialises a manifest to YAML bytes.
func ToYAML(m *Manifest) ([]byte, error) {
	return yaml.Marshal(m)
}
