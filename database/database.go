package database

import (
	"context"
	"embed"
	"fmt"
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
	"io/fs"
	"net/http"
	"nir/config"
	"nir/di"
)

//go:embed migrations/*.sql
var migrationsPath embed.FS

type Database struct {
	connection *sqlx.DB
}

func New() *Database {
	return &Database{}
}

func (db *Database) Connect(ctx context.Context) error {
	return di.FromContext(ctx).Invoke(func(cfg *config.Config) error {
		if cfg.Database.Clean {
			err := db.migrate(cfg, migrate.Down)
			if err != nil {
				return err
			}
		}

		err := db.migrate(cfg, migrate.Up)
		if err != nil {
			return err
		}

		return db.connect(ctx, cfg, "%s:%s@tcp(%s)/%s?parseTime=true")
	})
}

func (db *Database) migrate(cfg *config.Config, direction migrate.MigrationDirection) error {
	migrationsDirectory, err := fs.Sub(migrationsPath, "migrations")
	if err != nil {
		return err
	}

	migrations := &migrate.HttpFileSystemMigrationSource{FileSystem: http.FS(migrationsDirectory)}

	_, err = migrate.Exec(db.connection.DB, cfg.Database.Driver, migrations, direction)

	return err
}

func (db *Database) connect(ctx context.Context, cfg *config.Config, dsn string) (err error) {
	db.connection, err = sqlx.ConnectContext(ctx, cfg.Database.Driver, fmt.Sprintf(dsn, cfg.Database.User, cfg.Database.Password, cfg.Database.Address, cfg.Database.Name))
	if err != nil {
		return fmt.Errorf("failed to connect db: %w", err)
	}

	db.connection.Close()

	return nil
}
