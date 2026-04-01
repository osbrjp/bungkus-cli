package configdata

import "embed"

//go:embed bases/*.json extras/*.json all:templates
var FS embed.FS
