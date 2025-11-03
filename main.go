package main

import (
	"embed"
	"github.com/dillonlara115/barracuda/cmd"
)

//go:embed web/dist
var frontendFiles embed.FS

func main() {
	// Pass embedded frontend files to cmd package
	cmd.SetFrontendFiles(frontendFiles)
	cmd.Execute()
}

