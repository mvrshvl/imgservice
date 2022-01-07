package database

import (
	"context"
	"embed"
	"fmt"
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
	"net/http"
	"nir/config"
	"nir/di"
	"strings"
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
			err := db.recreate(ctx, cfg)
			if err != nil {
				return err
			}
		}

		err := db.migrate(ctx)
		if err != nil {
			return err
		}

		return db.connect(ctx, cfg, "%s:%s@tcp(%s)/%s?parseTime=true")
	})
}

func (db *Database) recreate(ctx context.Context, cfg *config.Config) error {
	err := db.connect(ctx, cfg, "%s:%s@tcp(%s)/")
	if err != nil {
		return err
	}

	_, err = db.connection.Query("DROP DATABASE ?", cfg.Name)
	if err != nil {
		return err
	}

	_, err = db.connection.Query("CREATE DATABASE ?", cfg.Name)

	return err
}

func (db *Database) migrate(ctx context.Context) error {
	migrations := &migrate.HttpFileSystemMigrationSource{FileSystem: http.FS(dir)}

	_, err := migrate.Exec(db.connection.DB, cfg.GetDatabase().DriverName, migrations, migrate.Up)

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
