// Package cmd contains commands
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	version = "unknown"

	cfgFile string

	rootCmd = &cobra.Command{
		Use:   "nurd-commerce",
		Short: "Start commerce API",
	}
)

// Execute root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config.yaml", "config file (default is config.yaml)")
}
