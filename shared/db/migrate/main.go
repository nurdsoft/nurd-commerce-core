// Package main is db/migrate app
package main

import (
	"os"

	"github.com/nurdsoft/nurd-commerce-core/shared/db/migrate/migrate"
)

func main() {
	if err := migrate.Command.Execute(); err != nil {
		os.Exit(1)
	}
}
