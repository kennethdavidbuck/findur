// Package buildinfo owns the immutable source revision embedded in the API binary.
package buildinfo

import (
	"fmt"
	"regexp"
)

var (
	// SHA is replaced with -ldflags at build time. Development builds remain explicit.
	SHA     = "development"
	fullSHA = regexp.MustCompile(`^[0-9a-f]{40}$`)
)

// ValidateProduction rejects images that were not built from an exact Git revision.
func ValidateProduction() error {
	if !fullSHA.MatchString(SHA) {
		return fmt.Errorf("build SHA must be a 40-character lowercase hexadecimal Git revision")
	}
	return nil
}
