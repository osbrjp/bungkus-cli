package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

const (
	// Repo is the GitHub repository releases are published to.
	Repo = "osbrjp/bungkus-cli"
	// InstallScriptURL is the installer bungkus-cli update re-runs; it owns
	// asset resolution, checksum verification, and the atomic replace.
	InstallScriptURL = "https://raw.githubusercontent.com/" + Repo + "/main/install.sh"

	latestReleaseAPI = "https://api.github.com/repos/" + Repo + "/releases/latest"
	updateCheckTTL   = 24 * time.Hour
)

// DevVersion is the version a build carries when no ldflags were passed.
const DevVersion = "dev"

// NormalizeVersion makes a version string comparable with semver: release tags
// are published as "vX.Y.Z", but a build may report "X.Y.Z".
func NormalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	if v != "" && !strings.HasPrefix(v, "v") {
		v = "v" + v
	}
	return v
}

// IsNewer reports whether latest is a newer release than current. An
// unparseable version on either side (a "dev" build, an odd tag) is never
// newer — the caller decides what to do about that, silence is not an upgrade.
func IsNewer(current, latest string) bool {
	c, l := NormalizeVersion(current), NormalizeVersion(latest)
	if !semver.IsValid(c) || !semver.IsValid(l) {
		return false
	}
	return semver.Compare(l, c) > 0
}

// LatestRelease returns the newest published release tag, e.g. "v1.4.0".
func LatestRelease(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, latestReleaseAPI, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github returned %s for the latest release", resp.Status)
	}

	var body struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}
	if body.TagName == "" {
		return "", fmt.Errorf("no published release found for %s", Repo)
	}
	return body.TagName, nil
}

// CachedLatest returns the latest release tag, fetching it only when the cache
// at path is older than ttl. The freshly fetched tag is written back with the
// file's mtime as the check timestamp. A cache that cannot be read or written
// costs an extra fetch, never an error.
func CachedLatest(path string, now time.Time, ttl time.Duration, fetch func() (string, error)) (string, error) {
	if info, err := os.Stat(path); err == nil && now.Sub(info.ModTime()) < ttl {
		if tag, err := os.ReadFile(path); err == nil && len(tag) > 0 {
			return strings.TrimSpace(string(tag)), nil
		}
	}

	tag, err := fetch()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err == nil {
		_ = os.WriteFile(path, []byte(tag), 0o644)
	}
	return tag, nil
}

// UpdateCachePath is where the once-a-day check records what it last saw.
func UpdateCachePath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "bungkus-cli", "latest-release"), nil
}

// AvailableUpdate returns the newest release tag when it is newer than
// current, or "" when current is up to date, is a dev build, or the check
// could not run. Network failures are not the user's problem here.
func AvailableUpdate(current string) string {
	if current == "" || current == DevVersion {
		return ""
	}
	path, err := UpdateCachePath()
	if err != nil {
		return ""
	}
	latest, err := CachedLatest(path, time.Now(), updateCheckTTL, func() (string, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		return LatestRelease(ctx)
	})
	if err != nil || !IsNewer(current, latest) {
		return ""
	}
	return latest
}
