/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>

*/
package main

import (
	"fmt"
	"os"

	"github.com/spencer-osbrjp/bungkus-cli/cmd"
	"github.com/spencer-osbrjp/bungkus-cli/config"
	"github.com/spencer-osbrjp/bungkus-cli/pkg"
)

func main() {
	if err := pkg.InitRegistry(config.RegistryJSON); err != nil {
		fmt.Fprintf(os.Stderr, "failed to load registry: %v\n", err)
		os.Exit(1)
	}
	cmd.SetVersion(Version)
	cmd.Execute()
}
