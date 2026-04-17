// Package platform centralises OS and architecture detection.
//
// Armada identifies a runtime environment by its Go-style "<os>/<arch>"
// pair (e.g. "linux/amd64", "darwin/arm64"). Manifests use the same
// convention for their `assets` map keys.
package platform

import (
	"fmt"
	"runtime"
)

// Info describes the running platform.
type Info struct {
	OS   string // "linux", "darwin", "windows", ...
	Arch string // "amd64", "arm64", ...
}

// Detect returns Info for the running process.
func Detect() Info {
	return Info{OS: runtime.GOOS, Arch: runtime.GOARCH}
}

// Key returns the canonical "<os>/<arch>" string used as the manifest
// assets map key.
func (i Info) Key() string {
	return fmt.Sprintf("%s/%s", i.OS, i.Arch)
}

// String implements fmt.Stringer.
func (i Info) String() string {
	return i.Key()
}

// ExecSuffix returns the extension expected for native executables on the
// current platform (".exe" on Windows, "" elsewhere).
func (i Info) ExecSuffix() string {
	if i.OS == "windows" {
		return ".exe"
	}
	return ""
}
