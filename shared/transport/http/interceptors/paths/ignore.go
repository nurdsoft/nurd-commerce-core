// Package paths provides ways to check specifics about HTTP request paths
package paths

import (
	"regexp"
)

var (
	ignoredPaths = regexp.MustCompile(`metrics|health|swaggerui|ws`)
)

// IsIgnoredPath tells us if the path needs to be ignored from further processing.
func IsIgnoredPath(path string) bool {
	return ignoredPaths.MatchString(path)
}
