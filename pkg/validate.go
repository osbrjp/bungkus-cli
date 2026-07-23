package pkg

import (
	"fmt"
	"path/filepath"
	"regexp"
)

// projectNameRe is the safe subset of npm package-name rules: a lowercase
// alphanumeric first character followed by lowercase alphanumerics and the
// separators '.', '-', '_'. Because it forbids '/' and a leading '.', a value
// that matches is always a single, local path segment — safe to use directly
// as the scaffold destination directory (no traversal, no absolute path).
var projectNameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*$`)

// maxProjectNameLen mirrors npm's 214-character package-name cap; the project
// name is emitted verbatim as package.json "name".
const maxProjectNameLen = 214

// ValidateProjectName rejects names that are not valid npm package names. This
// doubles as the traversal/absolute-path guard for the explicit `create <name>`
// path, since the name becomes the destination directory.
func ValidateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name must not be empty")
	}
	if len(name) > maxProjectNameLen {
		return fmt.Errorf("project name %q is too long (max %d characters)", name, maxProjectNameLen)
	}
	if !projectNameRe.MatchString(name) {
		return fmt.Errorf("invalid project name %q: use lowercase letters, digits, '.', '-', '_', and start with a letter or digit", name)
	}
	return nil
}

// ValidateDest ensures the resolved destination stays within the current
// working directory — no "..", no absolute path, not empty. "." (scaffold into
// the current directory) is local and allowed. This is a defense-in-depth check
// for every scaffold entry point (CLI arg, "." handling, and the TUI).
func ValidateDest(destDir string) error {
	if destDir == "." {
		return nil // scaffold-in-place into the current directory
	}
	if !filepath.IsLocal(destDir) {
		return fmt.Errorf("destination %q must be within the current directory (no absolute paths or '..')", destDir)
	}
	return nil
}
