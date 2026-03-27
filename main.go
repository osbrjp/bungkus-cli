package main

import (
	_ "embed"

	"github.com/spencer-osbrjp/bungkus-cli/cmd"
	"github.com/spencer-osbrjp/bungkus-cli/internal/patcher"
)

//go:embed patcher/dist/index.js
var patcherJS []byte

func main() {
	patcher.SetJS(patcherJS)
	cmd.Execute()
}
