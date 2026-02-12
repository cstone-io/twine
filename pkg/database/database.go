package database

import (
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/cstone-io/twine/pkg/config"
	"github.com/cstone-io/twine/pkg/errors"
	"github.com/cstone-io/twine/pkg/logger"
)

var instance *Database

// Database provides singleton access to GORM
type Database struct {
	mu         sync.Mutex
	client     *gorm.DB
	migrations []*Migration
}

// Get returns the singleton database instance
func Get() *Database {
	if instance == nil {
		cfg := config.Get().Database
		initialize(cfg)
	}

	if instance == nil {
		logger.Get().Critical("Database instance is not initialized")
	}

	return instance
}

// GORM returns the underlying GORM client
func GORM() *gorm.DB {
	return Get().client
}

func initialize(cfg config.DatabaseConfig) *Database {
	log := logger.Get()

	dsn := cfg.DSN()

	client, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.CustomError(errors.ErrDatabaseConn.Wrap(err))
		return nil
	}

	// Enable the UUID extension
	client.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")

	instance = &Database{
		client:     client,
		migrations: migrations,
	}

	if err := instance.migrate(); err != nil {
		log.CustomError(errors.ErrDatabaseMigration.Wrap(err))
	}

	return instance
}

// RegisterMigration adds a migration to the database
func RegisterMigration(m *Migration) {
	migrations = append(migrations, m)
}

// RegisterMigrations adds multiple migrations to the database
func RegisterMigrations(ms ...*Migration) {
	migrations = append(migrations, ms...)
}

func (d *Database) migrate() error {
	sorted := []*Migration{}
	visited := make(map[string]bool)

	var visit func(*Migration) error
	visit = func(m *Migration) error {
		if visited[m.Name] {
			return nil
		}

		visited[m.Name] = true

		for _, dep := range m.Deps {
			if err := visit(dep); err != nil {
				return errors.ErrSortMigrations.Wrap(err).WithValue("dependency " + dep.Name + " of model " + m.Name)
			}
		}

		sorted = append(sorted, m)
		return nil
	}

	for _, migration := range d.migrations {
		if err := visit(migration); err != nil {
			return err
		}
	}

	d.migrations = sorted

	d.mu.Lock()
	defer d.mu.Unlock()

	for _, m := range d.migrations {
		if err := d.client.AutoMigrate(m.Model); err != nil {
			return errors.ErrMigrateTable.Wrap(err).WithValue("model " + m.Name)
		}
		logger.Get().Debug("Migrated table: %s", m.Name)
	}
	return nil
}
