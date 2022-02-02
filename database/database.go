package database

import (
	"context"
	"embed"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
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
		err := db.connect(ctx, cfg, "%s:%s@tcp(%s)/%s?parseTime=true")
		if err != nil {
			return err
		}

		if cfg.Database.Clean {
			err := db.migrate(cfg, migrate.Down)
			if err != nil {
				return err
			}
		}

		return db.migrate(cfg, migrate.Up)
	})
}

func (db *Database) GetConnection() *sqlx.DB {
	return db.connection
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

	return nil
}
