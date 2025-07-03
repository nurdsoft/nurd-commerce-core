package testutils

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/nurdsoft/nurd-commerce-core/shared/db"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type TestDBConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
	SSLMode  string
}

func DefaultTestDBConfig() *TestDBConfig {
	return &TestDBConfig{
		Host:     "localhost",
		Port:     5453,
		Database: "commerce-core",
		User:     "db",
		Password: "123",
		SSLMode:  "disable",
	}
}

type TestDBInstance struct {
	sqlDB  *sql.DB
	gormDB *gorm.DB
}

func SetupTestDB(t *testing.T, config *TestDBConfig) *TestDBInstance {
	if config == nil {
		config = DefaultTestDBConfig()
	}

	dbConn, gormDB, err := db.New(&db.Config{
		Postgres: db.Postgres{
			Host:     config.Host,
			Port:     config.Port,
			Database: config.Database,
			User:     config.User,
			Password: config.Password,
			SSLMode:  config.SSLMode,
		},
	})
	require.NoError(t, err)

	// Run migrations on test database
	migrations := &migrate.FileMigrationSource{Dir: "../../../migrations"}
	n, err := migrate.Exec(dbConn, "postgres", migrations, migrate.Up)
	require.NoError(t, err)
	t.Logf("Applied %d migrations to test database", n)

	return &TestDBInstance{
		sqlDB:  dbConn,
		gormDB: gormDB,
	}
}

func (tdb *TestDBInstance) CleanupTestDB() {
	if tdb.sqlDB != nil {
		tdb.sqlDB.Close()
	}
}

func (tdb *TestDBInstance) GetSQLDB() *sql.DB {
	return tdb.sqlDB
}

func (tdb *TestDBInstance) GetGormDB() *gorm.DB {
	return tdb.gormDB
}

func CreateTestContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}
