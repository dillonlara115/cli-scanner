package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	debug bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "baracuda",
	Short: "A fast, lightweight SEO website crawler",
	Long: `baracuda is a CLI tool for crawling websites and extracting SEO data.
It recursively crawls websites, extracts key SEO elements like titles, meta descriptions,
headings, and links, and exports the data to CSV or JSON format.`,
	Version: "1.0.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	// Display banner for main commands (not for --help, --version, etc.)
	if len(os.Args) > 1 {
		firstArg := os.Args[1]
		// Only show banner for actual commands, not flags
		if firstArg != "--help" && firstArg != "-h" && firstArg != "--version" && firstArg != "-v" && firstArg != "help" {
			displayBanner()
		}
	} else {
		// No args, show banner
		displayBanner()
	}

	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")
}

