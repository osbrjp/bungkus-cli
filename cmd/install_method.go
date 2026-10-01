package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/osbrjp/bungkus-cli/pkg"
)

// installMethod is how the running binary was installed, which decides how
// it updates itself.
type installMethod int

const (
	// methodRelease is a release binary from install.sh (or a local build):
	// updates re-run install.sh.
	methodRelease installMethod = iota
	// methodGo is a `go install github.com/osbrjp/bungkus-cli@<version>`
	// build: updates re-run go install so the copy in GOBIN stays the one
	// that is updated.
	methodGo
)

// installedBy is the running binary's install method, set by SetVersion.
var installedBy = methodRelease

// goModule is the module path go install takes.
const goModule = "github.com/" + pkg.Repo

// resolveVersion picks the version the binary reports and how it was
// installed. A version stamped at build time (-ldflags, as releases do)
// wins. Without one, the module version Go records in the binary is used:
// `go install …@latest` records the release tag (methodGo), while a local
// `go build` records "(devel)" and stays a dev build.
//
// bi is the result of debug.ReadBuildInfo; ok reports whether it was
// available.
func resolveVersion(stamped string, bi *debug.BuildInfo, ok bool) (string, installMethod) {
	if stamped != "" && stamped != pkg.DevVersion {
		return stamped, methodRelease
	}
	if ok && bi != nil && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version, methodGo
	}
	return pkg.DevVersion, methodRelease
}

// updateCommand is the command that updates a binary installed by m, as
// shown in update hints.
func updateCommand(m installMethod) string {
	if m == methodGo {
		return "go install " + goModule + "@latest"
	}
	return "bungkus-cli update"
}

// copiesOnPath lists every executable named bungkus-cli in the PATH list
// path, in lookup order (the first one is the one the shell runs). Entries
// that resolve to the same file (symlinks, duplicate PATH folders) are listed
// once.
func copiesOnPath(path string) []string {
	var found []string
	seen := map[string]bool{}
	for _, dir := range filepath.SplitList(path) {
		if dir == "" {
			continue
		}
		p := filepath.Join(dir, "bungkus-cli")
		info, err := os.Stat(p)
		if err != nil || info.IsDir() || info.Mode().Perm()&0o111 == 0 {
			continue
		}
		real, err := filepath.EvalSymlinks(p)
		if err != nil {
			real = p
		}
		if seen[real] {
			continue
		}
		seen[real] = true
		found = append(found, p)
	}
	return found
}

// duplicateNote warns when bungkus-cli is installed more than once on PATH:
// only the first copy runs, so the others are stale and an update (which
// updates the running copy) never reaches them. self is the running
// binary's resolved path. It returns "" when there is a single copy.
func duplicateNote(self, path string) string {
	copies := copiesOnPath(path)
	if len(copies) < 2 {
		return ""
	}
	var b strings.Builder
	b.WriteString("note: bungkus-cli is installed more than once; the first one on PATH runs:\n")
	for i, p := range copies {
		var tags []string
		if i == 0 {
			tags = append(tags, "runs")
		}
		if real, err := filepath.EvalSymlinks(p); err == nil && real == self {
			tags = append(tags, "this one")
		}
		label := ""
		if len(tags) > 0 {
			label = " (" + strings.Join(tags, ", ") + ")"
		}
		fmt.Fprintf(&b, "  %s%s\n", p, label)
	}
	b.WriteString("remove the ones you don't use, e.g. rm " + copies[len(copies)-1] + "\n")
	return b.String()
}

// selfPath returns the running binary's path with symlinks resolved, or ""
// when it can't be determined.
func selfPath() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	if real, err := filepath.EvalSymlinks(exe); err == nil {
		return real
	}
	return exe
}
