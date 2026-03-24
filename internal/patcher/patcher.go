package patcher

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spencer-osbrjp/bungkus-cli/config"
)

var embeddedJS []byte

func SetJS(js []byte) {
	embeddedJS = js
}

// getScriptPath returns the path to the cached patcher script,
// extracting it from the embedded bytes if needed.
func getScriptPath() (string, error) {
	cacheDir, err := cacheDir()
	if err != nil {
		return "", err
	}

	scriptPath := filepath.Join(cacheDir, "patcher.mjs")
	hashPath := filepath.Join(cacheDir, "patcher.hash")

	// Compute hash of embedded JS
	sum := sha256.Sum256(embeddedJS)
	currentHash := hex.EncodeToString(sum[:])

	// Check if cached version is up to date
	if cachedHash, err := os.ReadFile(hashPath); err == nil {
		if string(cachedHash) == currentHash {
			return scriptPath, nil
		}
	}

	// Extract to cache
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache dir: %w", err)
	}
	if err := os.WriteFile(scriptPath, embeddedJS, 0644); err != nil {
		return "", fmt.Errorf("failed to write patcher: %w", err)
	}
	if err := os.WriteFile(hashPath, []byte(currentHash), 0644); err != nil {
		return "", fmt.Errorf("failed to write patcher hash: %w", err)
	}

	return scriptPath, nil
}

func cacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home dir: %w", err)
	}
	return filepath.Join(home, ".bungkus", "cache"), nil
}

func run(args []string) error {
	scriptPath, err := getScriptPath()
	if err != nil {
		return err
	}

	nodeArgs := append([]string{scriptPath}, args...)
	cmd := exec.Command("node", nodeArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// PatchBaseConfig injects imports and plugins into a base framework config file.
func PatchBaseConfig(configPath string, imports, plugins []string) error {
	args := []string{configPath}

	if len(imports) > 0 {
		importsJSON, _ := json.Marshal(imports)
		args = append(args, "--imports", string(importsJSON))
	}
	if len(plugins) > 0 {
		pluginsJSON, _ := json.Marshal(plugins)
		args = append(args, "--plugins", string(pluginsJSON))
	}

	return run(args)
}

// PatchExtra patches an extra's config file with framework-specific settings.
func PatchExtra(configPath string, base config.ExtraBase) error {
	needsPatch := len(base.Imports) > 0 || len(base.Spreads) > 0 || len(base.JsonMerge) > 0
	if !needsPatch {
		return nil
	}

	args := []string{configPath}

	if len(base.Imports) > 0 {
		importsJSON, _ := json.Marshal(base.Imports)
		args = append(args, "--imports", string(importsJSON))
	}
	if len(base.Spreads) > 0 {
		spreadsJSON, _ := json.Marshal(base.Spreads)
		args = append(args, "--spreads", string(spreadsJSON))
	}
	if len(base.JsonMerge) > 0 {
		mergeJSON, _ := json.Marshal(base.JsonMerge)
		args = append(args, "--json-merge", string(mergeJSON))
	}

	return run(args)
}
