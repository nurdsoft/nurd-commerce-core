package migrate

import (
	"time"

	"github.com/nurdsoft/nurd-commerce-core/shared/cfg"
	"github.com/nurdsoft/nurd-commerce-core/shared/db"
	"github.com/nurdsoft/nurd-commerce-core/shared/log"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/spf13/cobra"
)

var (
	cfgFile          string
	migrationsFolder string
	config           Config
	direction        string
)

var Command = &cobra.Command{
	Use:          "migrate",
	Short:        "migrate database",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger, err := log.New()
		if err != nil {
			return errors.Wrap(err, "init logs failed")
		}

		err = cfg.Init("config", cfgFile, &config)
		if err != nil {
			return errors.Wrap(err, "init configs failed")
		}

		pg, _, err := db.New(&config.DB)
		if err != nil {
			return errors.Wrap(err, "failed to init postgresql client")
		}

		logger.Info("Starting migrations...")

		errs := make(chan error, 1)

		go func() {
			defer close(errs)

			migrations := &migrate.FileMigrationSource{Dir: migrationsFolder}
			var n int
			if direction == "down" {
				logger.Info("Rolling back most recent migration")
				// by default, we only want to rollback 1 migration
				n, err = migrate.ExecMax(pg, "postgres", migrations, migrate.Down, 1)
			} else {
				logger.Info("Applying migrations")
				n, err = migrate.Exec(pg, "postgres", migrations, migrate.Up)
			}
			if err != nil {
				errs <- err
				return
			}

			logger.Infof("Applied migrations: %d", n)
		}()

		select {
		case err := <-errs:
			if err != nil {
				return errors.Wrap(err, "failed to run migrations")
			}
		case <-time.After(10 * time.Minute):
			return errors.New("failed to run migrations, timeout after 10 mins")
		}

		return nil
	},
}

//nolint:gochecknoinits
func init() {
	Command.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is config.yaml)")
	Command.PersistentFlags().StringVar(&migrationsFolder, "migrations-dir", "migrations", "directory with migrations files")
	Command.PersistentFlags().StringVar(&direction, "direction", "up", "migration direction (up or down)")
}
