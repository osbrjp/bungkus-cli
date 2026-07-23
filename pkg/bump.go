package pkg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

// DefaultSoak is how long a release must have been published before `bump`
// will adopt it. Bleeding-edge releases are where regressions and pulled
// versions live; a soak period lets the ecosystem shake those out first.
const DefaultSoak = 14 * 24 * time.Hour // 14 days

// Packument is the subset of an npm registry document that `bump` reads.
// deprecated is left raw because npm serves it as a message string when a
// version is deprecated but a bool (false) on some packages (e.g. react).
type Packument struct {
	Versions map[string]struct {
		Deprecated json.RawMessage `json:"deprecated"`
	} `json:"versions"`
	Time map[string]string `json:"time"` // version -> RFC3339 publish time
}

// isDeprecated reports whether a raw npm `deprecated` value marks the version
// unsafe: any non-empty message string, or an explicit true.
func isDeprecated(raw json.RawMessage) bool {
	switch strings.TrimSpace(string(raw)) {
	case "", "null", "false", `""`:
		return false
	default:
		return true
	}
}

// PickVersion returns the newest version of a package that is safe to adopt:
// a stable (non-prerelease) release, not deprecated, and published at least
// minAge before now. ok is false when nothing qualifies. The returned string
// is a bare version with no range prefix.
func PickVersion(doc Packument, now time.Time, minAge time.Duration) (version string, ok bool) {
	cutoff := now.Add(-minAge)
	best := ""
	for ver, meta := range doc.Versions {
		sv := "v" + ver
		if !semver.IsValid(sv) || semver.Prerelease(sv) != "" {
			continue // skip malformed and pre-releases (alpha/beta/rc/next)
		}
		if isDeprecated(meta.Deprecated) {
			continue // maintainer flagged it unsafe
		}
		ts, has := doc.Time[ver]
		if !has {
			continue
		}
		pub, err := time.Parse(time.RFC3339, ts)
		if err != nil || pub.After(cutoff) {
			continue // unparseable, or still inside the soak window
		}
		if best == "" || semver.Compare(sv, "v"+best) > 0 {
			best = ver
		}
	}
	return best, best != ""
}

// FetchPackument retrieves a package's npm registry document.
func FetchPackument(client *http.Client, name string) (Packument, error) {
	var doc Packument
	resp, err := client.Get("https://registry.npmjs.org/" + name)
	if err != nil {
		return doc, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return doc, fmt.Errorf("npm returned %s for %s", resp.Status, name)
	}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return doc, err
	}
	return doc, nil
}

// BumpChange records a single version pin that bump would rewrite.
type BumpChange struct {
	Name string
	From string // raw value including range prefix, e.g. "^6.1.3"
	To   string // rewritten value with the same prefix, e.g. "^6.4.0"
}

// BumpResult is the outcome of a dry run or write.
type BumpResult struct {
	Changes []BumpChange
	Skipped []string // package names no version could be resolved for
	Content string   // the rewritten registry.json content
}

// leading run of range operators (^, ~, >=, etc.) before the numeric version.
var rangePrefix = regexp.MustCompile(`^[\^~><=\s]*`)

// BumpRegistry parses a registry.json body, asks resolve() for the target bare
// version of every pinned package (resolve returns ok=false to leave one
// untouched), and rewrites the pins in place — preserving each pin's range
// prefix and the rest of the file byte-for-byte, so diffs stay minimal.
func BumpRegistry(content string, resolve func(name string) (string, bool)) (BumpResult, error) {
	var reg Registry
	if err := json.Unmarshal([]byte(content), &reg); err != nil {
		return BumpResult{}, fmt.Errorf("parse registry: %w", err)
	}

	// Collect every distinct (name, rawValue) pin across the whole registry.
	pins := map[string]map[string]bool{} // name -> set of raw values
	record := func(p Packages) {
		for n, v := range p.Dependencies {
			if pins[n] == nil {
				pins[n] = map[string]bool{}
			}
			pins[n][v] = true
		}
		for n, v := range p.DevDependencies {
			if pins[n] == nil {
				pins[n] = map[string]bool{}
			}
			pins[n][v] = true
		}
	}
	record(reg.CommonPackages)
	for _, b := range reg.Bases {
		record(b.Packages)
	}
	for _, group := range [][]OptionEntry{
		reg.CSS, reg.Formatters, reg.Linters, reg.Validation, reg.Form,
		reg.Query, reg.State, reg.CMS, reg.Test, reg.Audit, reg.Deployment, reg.CICD,
	} {
		for _, e := range group {
			record(e.Packages)
			for _, ip := range e.IntegrationPackages {
				record(ip)
			}
		}
	}

	res := BumpResult{Content: content}
	seen := map[string]bool{} // dedupe change reporting
	for name, values := range pins {
		target, ok := resolve(name)
		if !ok || target == "" {
			res.Skipped = append(res.Skipped, name)
			continue
		}
		for raw := range values {
			prefix := rangePrefix.FindString(raw)
			next := prefix + target
			if next == raw {
				continue
			}
			key := name + "\x00" + raw
			if seen[key] {
				continue
			}
			seen[key] = true

			// Match this exact pin only (scripts share package names but not
			// version-shaped values), tolerating whitespace around the colon.
			pat := regexp.MustCompile(`"` + regexp.QuoteMeta(name) + `"(\s*:\s*)"` + regexp.QuoteMeta(raw) + `"`)
			res.Content = pat.ReplaceAllString(res.Content, `"`+name+`"${1}"`+next+`"`)
			res.Changes = append(res.Changes, BumpChange{Name: name, From: raw, To: next})
		}
	}
	return res, nil
}
