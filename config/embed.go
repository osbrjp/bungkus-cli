package config

import "embed"

//go:embed all:templates
var Templates embed.FS

//go:embed registry.json
var RegistryJSON []byte
