package seed

import (
	"context"
	"embed"
	"io/fs"
	"text/template"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dbfixture"

	"git.dev.bi.zone/pam/onprem/dev/tools/review-flow-bot/internal/db/models"
)

//go:embed *.yaml
var fixturesFS embed.FS

// Run seeds the database with test data from YAML fixtures
func Run(ctx context.Context, db *bun.DB) error {
	log.Info().Msg("seeding database from fixtures...")

	db.RegisterModel(
		(*models.User)(nil),
		(*models.Team)(nil),
		(*models.TeamMember)(nil),
		(*models.MergeRequest)(nil),
		(*models.MRReviewer)(nil),
	)

	funcMap := template.FuncMap{
		"now": func() string {
			return time.Now().Format(time.RFC3339Nano)
		},
		"now_sub": func(duration string) string {
			d, _ := time.ParseDuration(duration)
			return time.Now().Add(-d).Format(time.RFC3339Nano)
		},
	}

	fixture := dbfixture.New(db, dbfixture.WithTemplateFuncs(funcMap))
	dir, err := fixturesFS.ReadDir(".")
	if err != nil {
		return errors.Wrap(err, "read embedded dir")
	}

	files := lo.Map(dir, func(item fs.DirEntry, _ int) string {
		return item.Name()
	})

	if err := fixture.Load(ctx, fixturesFS, files...); err != nil {
		return err
	}

	log.Info().Msg("database seeding completed")
	return nil
}
