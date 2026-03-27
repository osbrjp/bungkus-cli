package configdata

import "embed"

//go:embed setup.json extras/*.json all:templates
var FS embed.FS
