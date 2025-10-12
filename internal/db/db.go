package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/samber/do/v2"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/migrate"
	_ "modernc.org/sqlite"

	"github.com/chaindead/review-flow-bot/internal/config"
	"github.com/chaindead/review-flow-bot/internal/db/migrations"
	"github.com/chaindead/review-flow-bot/internal/db/seed"
)

type DB struct {
	cfg config.DB `do:"cfg.db"`
	db  *bun.DB
}

func New(i do.Injector) (*DB, error) {
	d, err := do.InvokeStruct[*DB](i)
	if err != nil {
		return nil, fmt.Errorf("invoke db: %w", err)
	}

	sqldb, err := sql.Open("sqlite", d.cfg.File)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	d.db = bun.NewDB(sqldb, sqlitedialect.New())
	_, _ = d.db.Exec("PRAGMA foreign_keys = ON")

	if err := d.migrate(); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	if d.cfg.Seed {
		if err := d.seed(); err != nil {
			return nil, fmt.Errorf("seed database: %w", err)
		}
	}

	log.Info().Str("file", d.cfg.File).Msg("database initialized")

	return d, nil
}

func (d *DB) migrate() error {
	ctx := context.Background()
	migrator := migrate.NewMigrator(d.db, migrations.Migrations)

	if err := migrator.Init(ctx); err != nil {
		return fmt.Errorf("init migrator: %w", err)
	}

	if err := migrator.Lock(ctx); err != nil {
		return fmt.Errorf("lock migrations: %w", err)
	}
	defer migrator.Unlock(ctx)

	group, err := migrator.Migrate(ctx)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	if group.IsZero() {
		log.Info().Msg("no new migrations to run")
	} else {
		log.Info().
			Int64("id", group.ID).
			Int("count", len(group.Migrations.Applied())).
			Msg("migrations applied")
	}

	return nil
}

func (d *DB) seed() error {
	ctx := context.Background()
	if err := seed.Run(ctx, d.db); err != nil {
		return fmt.Errorf("run seed: %w", err)
	}
	return nil
}

func (d *DB) DB() *bun.DB {
	return d.db
}

func (d *DB) Shutdown() error {
	log.Info().Msg("closing database connection")
	if err := d.db.Close(); err != nil {
		return fmt.Errorf("close db: %w", err)
	}
	return nil
}
