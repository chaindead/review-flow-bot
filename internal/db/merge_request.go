package db

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/chaindead/review-flow-bot/internal/db/models"
)

func (d *DB) MergeReqCreate(ctx context.Context,
	mrID, pID int,
	tID int64,
	title, url string,
	reviewers []models.User,
) (models.MergeRequest, error) {
	mr := models.MergeRequest{
		MRID:      mrID,
		ProjectID: pID,
		AuthorID:  tID,
		Title:     title,
		Link:      url,
	}

	if _, err := d.db.NewInsert().Model(&mr).Exec(ctx); err != nil {
		log.Error().Err(err).Msg("failed to insert merge_request")
		return mr, err
	}

	var mrr []*models.MRReviewer
	for _, user := range reviewers {
		mrr = append(mrr, &models.MRReviewer{
			MRID:       mrID,
			ProjectID:  pID,
			ReviewerID: user.ID,
		})

	}

	if _, err := d.db.NewInsert().Model(&mrr).Exec(ctx); err != nil {
		log.Error().Err(err).Msg("failed to insert reviewers")
		return mr, err
	}

	if err := d.db.NewSelect().Model(&mr).WherePK().
		Relation("Reviewers.Reviewer").
		Relation("Author").
		Scan(ctx); err != nil {
		log.Error().Err(err).Msg("failed to get mr")
		return mr, err
	}

	return mr, nil

}
