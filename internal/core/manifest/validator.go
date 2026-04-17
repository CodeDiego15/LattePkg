package manifest

import (
	"fmt"
	"strings"

	"github.com/DiegoDev2/armada/pkg/manifest"
)

// ValidationError describes a single manifest validation failure.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Validate checks a manifest for obvious errors. A manifest is considered
// valid if:
//
//   - name and version are set,
//   - either the modern `assets` map or the legacy `platforms` section is
//     populated,
//   - every declared asset has a URL and a properly formatted checksum.
//
// Legacy platforms are only validated superficially — they are expected to
// round-trip through `armada simulate`, not through the real installer.
func Validate(m *manifest.Manifest) []ValidationError {
	var errs []ValidationError

	if m.Name == "" {
		errs = append(errs, ValidationError{Field: "name", Message: "is required"})
	}
	if m.Version == "" {
		errs = append(errs, ValidationError{Field: "version", Message: "is required"})
	}

	hasAssets := len(m.Assets) > 0
	hasPlatforms := len(m.Platforms) > 0
	if !hasAssets && !hasPlatforms {
		errs = append(errs, ValidationError{
			Field:   "assets",
			Message: "at least one asset (or legacy 'platforms' entry) is required",
		})
	}

	for key, asset := range m.Assets {
		prefix := fmt.Sprintf("assets.%s", key)
		if asset.URL == "" {
			errs = append(errs, ValidationError{Field: prefix + ".url", Message: "is required"})
		}
		if asset.Checksum == "" {
			errs = append(errs, ValidationError{Field: prefix + ".checksum", Message: "is required"})
		} else if !strings.Contains(asset.Checksum, ":") {
			errs = append(errs, ValidationError{
				Field:   prefix + ".checksum",
				Message: "must be in the format '<algo>:<hex>'",
			})
		}
		if !strings.Contains(key, "/") {
			errs = append(errs, ValidationError{
				Field:   prefix,
				Message: "key must be '<os>/<arch>'",
			})
		}
	}

	// Legacy platform validation kept light.
	for p, pd := range m.Platforms {
		if len(pd.Architecture) == 0 && len(pd.PackageManagers) == 0 && pd.Fallback == nil {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("platforms.%s", p),
				Message: "must specify at least one architecture, package manager, or fallback",
			})
		}
	}

	return errs
}
