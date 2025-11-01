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
headings, and links, and exports the data to CSV or JSON format.

When run without arguments, baracuda starts in interactive mode.`,
	Version: "1.0.0",
	Run: func(cmd *cobra.Command, args []string) {
		// When baracuda is run without subcommands, start interactive crawl
		displayBanner()
		fmt.Println()
		// Force interactive mode when called from root
		interactive = true
		// Run the crawl command with interactive mode
		if err := runCrawl(cmd, args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")
}

