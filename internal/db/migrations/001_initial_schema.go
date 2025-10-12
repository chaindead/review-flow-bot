package migrations

import (
	"context"
	"slices"

	"github.com/uptrace/bun"

	"github.com/chaindead/review-flow-bot/internal/db/models"
)

func init() {
	ms := []interface{}{
		(*models.User)(nil),
		(*models.Team)(nil),
		(*models.TeamMember)(nil),
		(*models.MergeRequest)(nil),
		(*models.MRReviewer)(nil),
	}

	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		for _, model := range ms {
			if _, err := db.NewCreateTable().
				Model(model).
				IfNotExists().
				Exec(ctx); err != nil {
				return err
			}
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		slices.Reverse(ms)
		for _, model := range ms {
			if _, err := db.NewDropTable().
				Model(model).
				IfExists().
				Exec(ctx); err != nil {
				return err
			}
		}

		return nil
	})
}
