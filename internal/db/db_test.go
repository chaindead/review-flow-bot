package db

import (
	"testing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun/extra/bundebug"

	"git.dev.bi.zone/pam/onprem/dev/tools/review-flow-bot/internal/config"
)

func getDB(t *testing.T) *DB {
	t.Helper()
	injector := do.New()

	testConfig := config.DB{
		File: ":memory:",
		Seed: true,
	}
	do.ProvideNamedValue(injector, "cfg.db", testConfig)
	do.Provide(injector, New)

	db := do.MustInvoke[*DB](injector)
	require.NotNil(t, db, "database should not be nil")

	db.DB().AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))

	t.Cleanup(func() {
		require.NoError(t, db.Shutdown())
	})

	return db
}
