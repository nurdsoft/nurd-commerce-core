package cmd

import "github.com/nurdsoft/nurd-commerce-core/shared/db/migrate/migrate"

func init() {
	rootCmd.AddCommand(migrate.Command)
}
